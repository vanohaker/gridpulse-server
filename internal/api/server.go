package api

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vanohaker/gridpulse-server/internal/config"
	"github.com/vanohaker/gridpulse-server/internal/database/postgres"
)

var ServerInterface interface {
	Livenesprobe(*fiber.Ctx) error
	DeviceAddV1(*fiber.Ctx) error
	AddOauthProviderV1(*fiber.Ctx) error
	UserRegisterV1(*fiber.Ctx) error
	LoginUserV1(*fiber.Ctx) error
	RefreshAcessTokenV1(*fiber.Ctx) error
}

type Server struct {
	Pgdb   *postgres.DatabaseStr
	Rdb    *redis.Client
	Logger zerolog.Logger
	Ctx    context.Context
	Conf   *config.ConfigYaml
}

func NewServer(server Server) Server {
	return server
}
