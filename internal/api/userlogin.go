package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanohaker/gridpulse-server/ogen"
)

// Это логин пользователя.
// Должно вернуть два токена acess и refresh и ещё информацию о пользователе
func (s Server) LoginUserV1(c *fiber.Ctx) error {
	reqData := new(ogen.LoginUserV1Req) // Выделяем память для данных которе пришли в запросе
	// Если запрос не рсспарсился структурой то ошибка
	if err := c.BodyParser(reqData); err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	// Контекст чтобы ждать не более 10 секунд чтобы прокерить логин
	ctx, cancel := context.WithTimeout(s.Ctx, time.Second*10)
	defer cancel()
	// Ищем пользователя в базе данных
	user, err := s.Pgdb.SearchUserByName(ctx, reqData.Username)
	// Если пользователь не найден
	if user == nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusNotFound,
			err:    errors.New("user not found"),
		})
	}
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	// Проверяем что пароль валидный если пользовтель есть
	err = verifyPassword(reqData.Password, user.PasswordHashed)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusForbidden,
			err:    errors.New("password not match"),
		})
	}
	// Создаём быстрый acesstoken
	acesstokentd := time.Minute * 15
	at, err := s.createtoken(user.Username, acesstokentd)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	// Содаём долгий refrashtoken
	refreshtokentd := time.Minute * 44640
	rt, err := s.createtoken(user.Username, refreshtokentd)
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	// Надо записать токены в redis
	ctx, cancel = context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	// Пишем acesstoken
	err = s.Rdb.Set(ctx, fmt.Sprintf("acesstoken-%s", user.Id), at, acesstokentd).Err()
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	err = s.Rdb.Set(ctx, fmt.Sprintf("refreshtoken-%s", user.Id), rt, refreshtokentd).Err()
	if err != nil {
		return loginResponde(c, respondeData{
			status: fiber.StatusInternalServerError,
			err:    err,
		})
	}
	// Если всё правильно то возвращяем данные ользователю
	return loginResponde(c, respondeData{
		status:       fiber.StatusOK,
		acesstoken:   at,
		refreshtoken: rt,
		username:     user.Username,
		email:        user.Email,
	})
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
	case fiber.StatusForbidden:
		return c.Status(fiber.StatusForbidden).JSON(ogen.AcessDenied{
			Data: ogen.Data{
				Msg: r.err.Error(),
			},
		})
	case fiber.StatusNotFound:
		return c.Status(fiber.StatusNotFound).JSON(ogen.UserNotFound{
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
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(ogen.InternalServerError{
			Data: ogen.Data{
				Msg: errors.New("unexpected error").Error(),
			},
		})
	}
}
