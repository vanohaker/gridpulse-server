package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanohaker/gridpulse-server/ogen"
)

func (s Server) UserRegisterV1(c *fiber.Ctx) error {
	reqData := new(ogen.RegisterNewUser)
	if err := c.BodyParser(reqData); err != nil {
		return registerResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	users, _ := s.Pgdb.SearchUserByName(s.Ctx, reqData.Username)
	if users != nil {
		return registerResponde(c, respondeData{
			status: fiber.StatusForbidden,
			err:    errors.New("user exists"),
		})
	}
	hash, err := hashPassword(reqData.Password)
	if err != nil {
		return registerResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	if err := s.Pgdb.AddUser(s.Ctx, reqData.Username, hash, reqData.Email, "base", "user"); err != nil {
		return registerResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}

	at, err := s.createtoken(s.Conf.Jwtsecret, time.Minute*15)
	if err != nil {
		return registerResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	rt, err := s.createtoken(s.Conf.Jwtsecret, time.Minute*44640)
	if err != nil {
		return registerResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}

	return c.Status(200).JSON(ogen.RegisterNewUserSucess{
		Data: ogen.UserAuthData{
			Acesstoken:   at,
			Refreshtoken: rt,
			Userdata: ogen.UserData{
				Username: reqData.Username,
				Email:    reqData.Email,
			},
		},
	})
}

func registerResponde(c *fiber.Ctx, r respondeData) error {
	switch r.status {
	case fiber.StatusAccepted:
		return c.Status(fiber.StatusNotFound).JSON(ogen.AcessDenied{
			Data: ogen.Data{
				Msg: r.err.Error(),
			},
		})
	case fiber.StatusInternalServerError:
		return c.Status(fiber.StatusInternalServerError).JSON(ogen.InternalServerError{
			Data: ogen.Data{
				Msg: r.err.Error(),
			},
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(ogen.InternalServerError{
		Data: ogen.Data{
			Msg: errors.New("unexpected error").Error(),
		},
	})
}
