package api

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func verifytoken(acesstoken string) (bool, error) {
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
