package api

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanohaker/gridpulse-server/ogen"
)

type livenesStatus struct {
	code     int
	postgres string
	redis    string
}

func (s Server) Livenesprobe(c *fiber.Ctx) error {
	var statuses = livenesStatus{
		code: fiber.StatusOK,
	}
	var wg sync.WaitGroup
	wg.Add(2)
	postgresChain := make(chan error, 1)
	redisChain := make(chan error, 1)
	pgctx, pgctxclose := context.WithTimeout(s.Ctx, time.Second*10)
	defer pgctxclose()
	go func() {
		defer wg.Done()
		select {
		case <-pgctx.Done():
			postgresChain <- pgctx.Err()
			return
		default:
			err := s.Pgdb.Ping(pgctx)
			if err != nil {
				s.Logger.Warn().Err(err).Msg("")
				postgresChain <- err
				return
			}
			postgresChain <- nil
		}
	}()
	rctx, tctxclose := context.WithTimeout(s.Ctx, time.Second*10)
	defer tctxclose()
	go func() {
		defer wg.Done()
		select {
		case <-rctx.Done():
			redisChain <- rctx.Err()
		default:
			_, err := s.Rdb.Ping(rctx).Result()
			if err != nil {
				s.Logger.Warn().Err(err).Msg("")
				redisChain <- err
				return
			}
			redisChain <- nil
		}
	}()
	go func() {
		wg.Wait()
		close(postgresChain)
		close(redisChain)
	}()
	for i := 0; i < 2; i++ {
		select {
		// postgres status
		case ps := <-postgresChain:
			if ps != nil {
				statuses.postgres = ps.Error()
				statuses.code = fiber.StatusInternalServerError
			} else {
				statuses.postgres = "OK"
			}
		case rs := <-redisChain:
			if rs != nil {
				statuses.redis = rs.Error()
				statuses.code = fiber.StatusInternalServerError
			} else {
				statuses.redis = "OK"
			}
		}

	}
	return c.Status(statuses.code).JSON(ogen.LivenesProbe{
		Data: ogen.LivenesProbeData{
			Postgres: statuses.postgres,
			Redis:    statuses.redis,
		},
	})
}
