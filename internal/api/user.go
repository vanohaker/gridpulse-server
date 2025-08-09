package api

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vanohaker/gridpulse-server/ogen"
	"golang.org/x/crypto/bcrypt"
)

type respondeData struct {
	// responde code
	status int
	// error
	err error
	// jwt acess token
	acesstoken string
	// jwt refresh token
	refreshtoken string
	username     string
	email        string
}

func loginResponde(c *fiber.Ctx, r respondeData) error {
	switch r.status {
	case fiber.StatusOK:
		return c.Status(fiber.StatusOK).JSON(ogen.LoginSucess{
			Data: ogen.UserAuthData{
				Acesstoken:   r.acesstoken,
				Refreshtoken: r.refreshtoken,
				Userdata: ogen.UserData{
					Username: r.username,
					Email:    r.email,
				},
			},
		})
	case fiber.StatusInternalServerError, fiber.StatusForbidden:
		return c.Status(fiber.StatusInternalServerError).JSON(ogen.InternalServerError{
			Data: ogen.Data{
				Msg: r.err.Error(),
			},
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(ogen.InternalServerError{
			Data: ogen.Data{
				Msg: errors.New("unexpected error").Error(),
			},
		})
	}
}

func regiusterResponde(c *fiber.Ctx, r respondeData) error {
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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s Server) verifytoken(acesstoken string) (bool, error) {
	return true, nil
}

func (s Server) createtoken(username string, exp time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(exp).Unix(),
		"jti": uuid.New(),
	})
	t, err := token.SignedString([]byte(s.Conf.Jwtsecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (s Server) UserRegisterV1(c *fiber.Ctx) error {
	reqData := new(ogen.RegisterNewUser)
	if err := c.BodyParser(reqData); err != nil {
		return regiusterResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	users, _ := s.Pgdb.SearchUserByName(s.Ctx, reqData.Username)
	if users != nil {
		return regiusterResponde(c, respondeData{
			status: fiber.StatusForbidden,
			err:    errors.New("user exists"),
		})
	}
	hash, err := hashPassword(reqData.Password)
	if err != nil {
		return regiusterResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	if err := s.Pgdb.AddUser(s.Ctx, reqData.Username, hash, reqData.Email, "base", "user"); err != nil {
		return regiusterResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}

	at, err := s.createtoken(s.Conf.Jwtsecret, time.Minute*15)
	if err != nil {
		return regiusterResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	rt, err := s.createtoken(s.Conf.Jwtsecret, time.Minute*44640)
	if err != nil {
		return regiusterResponde(c, respondeData{
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

func (s Server) LoginUserV1(c *fiber.Ctx) error {
	reqData := new(ogen.LoginUserV1Req)
	if err := c.BodyParser(reqData); err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	ctx, cancel := context.WithTimeout(s.Ctx, time.Second*10)
	defer cancel()
	user, err := s.Pgdb.SearchUserByName(ctx, reqData.Username)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	if user == nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusNotFound,
			err:    errors.New("user not found"),
		})
	}
	err = verifyPassword(reqData.Password, user.PasswordHashed)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusForbidden,
			err:    errors.New("password not match"),
		})
	}
	at, err := s.createtoken(user.Username, time.Minute*15)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	rt, err := s.createtoken(user.Username, time.Minute*44640)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	return loginResponde(c, respondeData{
		status:       fiber.StatusOK,
		acesstoken:   at,
		refreshtoken: rt,
		username:     user.Username,
		email:        user.Email,
	})
}
