package database

import (
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5/pgtype"
)

type Account struct {
	// UUID пользователя
	Id uuid.UUID `db:"id"`
	// Имя пользователя
	Username string `db:"username"`
	// Email пользователя
	Email string `db:"email"`
	// Таймстемп регистрации пользователя
	RegistrationDate pgtype.Timestamptz `db:"registration_date"`
	// Таймстемп редактирования записи в бд
	EditDate pgtype.Timestamptz `db:"edit_date"`
	// Хеш пароля
	PasswordHashed string `db:"password_hashed"`
	// Признак того что акаунт активен
	Enabled null.Bool `db:"enabled"`
	// Признак того что акаунт активирован
	Activated null.Bool `db:"activated"`
}
