package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vanohaker/gridpulse-server/ogen"
)

func (s Server) RefreshAcessTokenV1(c *fiber.Ctx) error {
	return c.Status(200).JSON(ogen.SucessRefreshToken{
		Data: ogen.Data{
			Msg: "Sucess",
		},
	})
}
