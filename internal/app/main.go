package app

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"cnmt/internal/common/env"
	"cnmt/internal/common/httpx"
	"cnmt/internal/features/auth"
	"cnmt/internal/features/countries"
	"cnmt/internal/features/dashboard"
	"cnmt/internal/features/paymentaccounts"
	"cnmt/internal/features/transfers"
	"cnmt/internal/infra/db"
	"cnmt/internal/infra/notifications"
	"cnmt/internal/infra/storage"
	"cnmt/internal/infra/workers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type AppConfig struct {
	dbConn *pgxpool.Pool
}

type App struct {
	Router *chi.Mux
	WorkerClient *river.Client[pgx.Tx]
	Logger *slog.Logger
}

func NewApp(dbConn *pgxpool.Pool) *AppConfig {
	return &AppConfig{
		dbConn: dbConn,
	}
}

func (app *AppConfig) Run() (*App, error) {
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
		return nil, fmt.Errorf("failed to initialize object storage: %v", err)
	}

	dbQueries := db.New(app.dbConn)

	notifier, err := notifications.NewFromEnv(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize notifications: %w", err)
	}

	workerFct := workers.NewWorkers(app.dbConn, notifier)
	workerClient, workerErr := workerFct.Init()

	if workerErr != nil {
		return nil, fmt.Errorf("failed to initialize workers: %v", workerErr)
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
		dbConn:       app.dbConn,
		dbQueries:    dbQueries,
		objStorage:   objStorage,
		logger:       logger,
		workerClient: workerClient,
	}

	initRoutes(r, config)

	return &App{Router: r, WorkerClient: workerClient, Logger: logger}, nil
}

type RoutesConfig struct {
	dbConn     *pgxpool.Pool
	dbQueries *db.Queries
	objStorage *storage.ObjStorage
	logger     *slog.Logger
	workerClient *river.Client[pgx.Tx]
}

func initRoutes(r *chi.Mux, config RoutesConfig) {

	jwtSecret := env.GetString("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		panic("auth: JWT_SECRET must be at least 32 characters")
	}

	tokenTTL := time.Duration(env.GetInt("JWT_ACCESS_TTL", 86400)) * time.Second
	jwtAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)

	authSvc := auth.NewService(config.dbQueries, config.logger, jwtAuth, tokenTTL)
	authMiddleware := auth.NewMiddleware(authSvc, jwtAuth)
	authCtrl := auth.NewController(authSvc)

	transferServiceConfig := transfers.ServiceConfig{
		DB:           config.dbConn,
		Queries:      config.dbQueries,
		ObjStorage:   config.objStorage,
		Logger:       config.logger,
		WorkerClient: config.workerClient,
	}
	transferCtrl := transfers.NewController(transfers.NewService(transferServiceConfig))
	countryCtrl := countries.NewController(countries.NewService(config.dbQueries, config.logger, config.dbConn))
	paymentAccountCtrl := paymentaccounts.NewController(paymentaccounts.NewService(config.dbQueries, config.logger))
	dashboardCtrl := dashboard.NewController(dashboard.NewService(config.dbQueries, config.logger))

	r.Route("/api/v1", func(r chi.Router) {
		transferCtrl.Routes(r)
		countryCtrl.Routes(r)
		paymentAccountCtrl.Routes(r)
		authCtrl.Routes(r)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Verifier())
			r.Use(authMiddleware.AuthenticateUser)
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireRole(auth.RoleAdmin))
				transferCtrl.AdminRoutes(r)
				countryCtrl.AdminRoutes(r)
				paymentAccountCtrl.AdminRoutes(r)
				dashboardCtrl.AdminRoutes(r)
				authCtrl.AuthenticatedRoutes(r)
			})
		})
	})
}
