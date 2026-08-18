package app

import (
	"log/slog"
	"os"
	"time"

	"cnmt/internal/common/env"
	"cnmt/internal/common/httpx"
	"cnmt/internal/features/countries"
	"cnmt/internal/features/transfers"
	"cnmt/internal/infra/db"
	"cnmt/internal/infra/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppConfig struct {
	dbConn *pgxpool.Pool
}

func NewApp(dbConn *pgxpool.Pool) *AppConfig {
	return &AppConfig{
		dbConn: dbConn,
	}
}

func (app *AppConfig) Run() (*chi.Mux, error) {
	r := chi.NewRouter()

	level := slog.LevelInfo
	if env.GetString("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	urlTTL := time.Duration(env.GetInt("OBJECT_STORAGE_PRESIGNED_URL_EXPIRATION", 3600)) * time.Second

	objStorage, err := storage.NewObjStorage(urlTTL)
	if err != nil {
		return nil, err
	}

	httpx.InitValidator()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.CleanPath)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	config := RoutesConfig{
		dbConn:     app.dbConn,
		objStorage: objStorage,
		logger:     logger,
	}

	initRoutes(r, config)

	return r, nil
}

type RoutesConfig struct {
	dbConn     *pgxpool.Pool
	objStorage *storage.ObjStorage
	logger     *slog.Logger
}

func initRoutes(r *chi.Mux, config RoutesConfig) {
	dbQueries := db.New(config.dbConn)

	transferCtrl := transfers.NewController(transfers.NewService(config.dbConn, dbQueries, config.objStorage, config.logger))
	countryCtrl := countries.NewController(countries.NewService(dbQueries, config.logger, config.dbConn))

	r.Route("/api/v1", func(r chi.Router) {
		transferCtrl.Routes(r)
		transferCtrl.AdminRoutes(r)
		countryCtrl.Routes(r)
		countryCtrl.AdminRoutes(r)
	})
}
