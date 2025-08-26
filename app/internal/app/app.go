package app

import (
	"app/internal/config"
	"app/internal/domain/product/storage"
	"app/pkg/client/postgresql"
	"app/pkg/client/postgresql/adapter"
	"app/pkg/common/logging"
	"app/pkg/metric"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	_ "app/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg        *config.Config
	router     *httprouter.Router
	httpServer *http.Server
	pgClient   *pgxpool.Pool
}

func NewApp(ctx context.Context, config *config.Config) (App, error) {
	logging.GetLogger(ctx).Info("router init")
	router := httprouter.New()

	logging.GetLogger(ctx).Info("swagger docs init")
	router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	logging.GetLogger(ctx).Info("heartbeat metric initializing")
	metricHandler := metric.Handler{}
	metricHandler.Register(router)

	pgConfig := postgresql.NewPgConfig(
		config.PostgresqSQL.Username, config.PostgresqSQL.Password,
		config.PostgresqSQL.Host, config.PostgresqSQL.Port, config.PostgresqSQL.Database,
	)

	pgClient, err := postgresql.NewClient(ctx, 5, time.Second*5, pgConfig)
	if err != nil {
		logging.GetLogger(ctx).Fatal(err)
	}

	db := adapter.New(pgClient) // ← оборачиваем пул
	productStorage := storage.NewProductStorage(db)
	all, err := productStorage.All(context.Background())
	if err != nil {
		logging.GetLogger(ctx).Fatal(err)
	}
	logging.GetLogger(ctx).Fatal(all)

	return App{
		cfg:      config,
		router:   router,
		pgClient: pgClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return a.startHTTP(ctx)
	})
	logging.GetLogger(ctx).Info("application init and start")
	return grp.Wait()
}

func (a *App) startHTTP(ctx context.Context) error {
	logging.GetLogger(ctx).WithFields(map[string]interface{}{
		"IP":   a.cfg.HTTP.IP,
		"Port": a.cfg.HTTP.Port,
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("%s: %d", a.cfg.HTTP.IP, a.cfg.HTTP.Port))
	if err != nil {
		logging.GetLogger(ctx).WithError(err).Fatal("failed to create listener")
	}

	logging.GetLogger(ctx).WithFields(map[string]interface{}{
		"AllowedMethods":     a.cfg.HTTP.CORS.AllowedMethods,
		"AllowedOrigins":     a.cfg.HTTP.CORS.AllowedOrigins,
		"AllowCredentials":   a.cfg.HTTP.CORS.AllowCredentials,
		"AllowedHeaders":     a.cfg.HTTP.CORS.AllowedHeaders,
		"OptionsPassthrough": a.cfg.HTTP.CORS.OptionsPassthrough,
		"ExposedHeaders":     a.cfg.HTTP.CORS.ExposedHeaders,
		"Debug":              a.cfg.HTTP.CORS.Debug,
	})

	c := cors.New(cors.Options{
		AllowedMethods:     a.cfg.HTTP.CORS.AllowedMethods,
		AllowedOrigins:     a.cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials:   a.cfg.HTTP.CORS.AllowCredentials,
		AllowedHeaders:     a.cfg.HTTP.CORS.AllowedHeaders,
		OptionsPassthrough: a.cfg.HTTP.CORS.OptionsPassthrough,
		ExposedHeaders:     a.cfg.HTTP.CORS.ExposedHeaders,
		Debug:              a.cfg.HTTP.CORS.Debug,
	})

	handler := c.Handler(a.router)

	a.httpServer = &http.Server{
		Handler:      handler,
		WriteTimeout: a.cfg.HTTP.WriteTimeout,
		ReadTimeout:  a.cfg.HTTP.ReadTimeout,
	}

	logging.GetLogger(ctx).Info("application completely initialized and started")

	if err := a.httpServer.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logging.GetLogger(ctx).Warningln("server shutdown")
		default:
			logging.GetLogger(ctx).Fatal(err)
		}
	}
	err = a.httpServer.Shutdown(context.Background())
	if err != nil {
		logging.GetLogger(ctx).Fatal(err)
	}
	return err
}
