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
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	_ "app/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
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

	pgClient, err := postgresql.NewClient(context.Background(), 5, time.Second*5, pgConfig)
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

func (a *App) Run(ctx context.Context) {
	a.startHTTP(ctx)
}

func (a *App) startHTTP(ctx context.Context) {
	logging.GetLogger(ctx).Info("start http")

	var listener net.Listener

	if a.cfg.Listen.Type == config.LISTEN_TYPE_SOCK {
		appDir, err := filepath.Abs((filepath.Dir(os.Args[0])))
		if err != nil {
			logging.GetLogger(ctx).Fatal(err)
		}
		socketPath := path.Join(appDir, a.cfg.Listen.SocketFile)

		log.Printf("socket path: %s", socketPath)
		log.Printf("create and listen unix socket")

		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logging.GetLogger(ctx).Fatal(err)
		}
	} else {
		log.Printf("bind application to host: %s and port: %s", a.cfg.Listen.BindIP, a.cfg.Listen.Port)
		var err error
		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", a.cfg.Listen.BindIP, a.cfg.Listen.Port))
		if err != nil {
			logging.GetLogger(ctx).Fatal(err)
		}
	}

	c := cors.New(cors.Options{
		AllowedMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions, http.MethodDelete},
		AllowedOrigins:     []string{"http://172.28.1.87:3000", "http://172.28.1.87:8080"},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"Location", "Charset", "Access-Control-Allow-Origin", "Content-Type", "content-type", "Origin", "Accept", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		OptionsPassthrough: true,
		ExposedHeaders:     []string{"Location", "Authorization", "Content-Disposition"},
		Debug:              false,
	})

	handler := c.Handler(a.router)

	a.httpServer = &http.Server{
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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
	err := a.httpServer.Shutdown(context.Background())
	if err != nil {
		logging.GetLogger(ctx).Fatal(err)
	}
}
