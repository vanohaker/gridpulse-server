//go:generate go tool oapi-codegen -config ../codegen.yaml ../api/openapi.yml
//go:generate go tool ogen --target ../ogen --package ogen ../api/openapi.yml

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vanohaker/gridpulse-server/codegen"
	"github.com/vanohaker/gridpulse-server/internal/api"
	"github.com/vanohaker/gridpulse-server/internal/config"
	"github.com/vanohaker/gridpulse-server/internal/database"
	_ "github.com/vanohaker/gridpulse-server/internal/migrations"
	glog "go.finelli.dev/gooseloggers/zerolog"
)

var (
	ctx      = context.Background()
	pgdb     *database.DatabaseStr
	rdb      *redis.Client
	conf     *config.ConfigYaml
	logger   zerolog.Logger
	migrate  *string
	confPath *string
	err      error
)

func init() {
	// Путь до конфигурационного файла
	confPath = flag.String("config", "config.yaml", "Configuration file")

	// Параметры миграции
	// up - накатить миграции до актуального состояния
	// down - откатить последнюю миграцию
	// status - статус миграций
	migrate = flag.String("migrate", "up", "Migration up/down")
	flag.Parse()

	// логер
	logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i any) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
	}).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()

	// параметры из конфигурационного файла
	conf, err = config.LoadConfig(logger, confPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	pgdb, err = database.Initialize(ctx, conf, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%v", conf.Redis.Host, conf.Redis.Port),
		DB:   conf.Redis.Database,
	})

}

func main() {
	defer pgdb.PgxPool.Close()
	defer rdb.Close()
	connection, err := pgdb.PgxPool.Acquire(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	err = connection.Ping(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	logger.Info().Msg("Connected to the database!!")

	// goose logic
	goose.SetLogger(glog.GooseZerologLogger(&logger)) // enable zerolog logger for goose
	db := sql.OpenDB(stdlib.GetPoolConnector(pgdb.PgxPool))
	goose.SetTableName("public.gridpulse_db_version")

	store, err := goosedb.NewStore(goosedb.DialectPostgres, "gridpulse_db_version")
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	gooseProvider, err := goose.NewProvider(
		"",
		db,
		os.DirFS("./internal/migrations"),
		goose.WithLogger(glog.GooseZerologLogger(&logger)),
		goose.WithStore(store),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}

	switch *migrate {
	case "up":
		status, err := gooseProvider.Up(ctx)
		if err != nil {
			logger.Fatal().Err(err).Msg("")
		}
		for _, s := range status {
			logger.Info().Msg(s.String())
		}
		db.Close()
	case "down":
		status, err := gooseProvider.Down(ctx)
		if err != nil {
			logger.Fatal().Err(err).Msg("")
		}
		logger.Info().Msg(status.String())
		os.Exit(0)
	case "status":
		status, err := gooseProvider.Status(ctx)
		if err != nil {
			logger.Fatal().Err(err).Msg("")
		}
		for _, s := range status {
			logger.Info().Msgf("%s : %s", s.State, s.Source.Path)
		}
		os.Exit(0)
	}

	server := api.NewServer(api.Server{
		Pgdb:   pgdb,
		Rdb:    rdb,
		Logger: logger,
		Ctx:    ctx,
		Conf:   conf,
	})
	app := fiber.New()
	cfg := swagger.Config{
		BasePath: "/",
		FilePath: "./api/openapi.yml",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}), swagger.New(cfg))

	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.SendString("/metrics")
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger")
	})

	codegen.RegisterHandlers(app, server)

	err = app.Listen(fmt.Sprintf("%s:%v", conf.AppRes.Bind, conf.AppRes.Port))
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
}
