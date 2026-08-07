package app

import (
	"time"

	"cnmt/internal/common/httpx"
	"cnmt/internal/features/countries"
	"cnmt/internal/features/transfers"
	"cnmt/internal/infra/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppConfig struct {
	dbConn *pgxpool.Pool
}

func NewApp(dbConn *pgxpool.Pool) (*AppConfig) {
	return &AppConfig{
		dbConn: dbConn,
	}
}

func (app *AppConfig) Run() *chi.Mux {
	r := chi.NewRouter()
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

	  initRoutes(r,app.dbConn)

	return r
}

func initRoutes(r *chi.Mux, dbConn *pgxpool.Pool) {
	dbQueries := db.New(dbConn)
	
	transferCtrl := transfers.NewController(transfers.NewService(dbConn, dbQueries))
	countryCtrl := countries.NewController(countries.NewService(dbQueries))

	r.Route("/api/v1", func(r chi.Router) {
		transferCtrl.Routes(r)
		countryCtrl.Routes(r)
	})
}