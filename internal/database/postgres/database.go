package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/vanohaker/gridpulse-server/internal/config"
)

type DatabaseStr struct {
	PgxPool *pgxpool.Pool
	conf    *config.ConfigYaml
	logger  zerolog.Logger
}

func dbConfig(conf *config.ConfigYaml) (*pgxpool.Config, error) {
	dbConfig, err := pgxpool.ParseConfig(conf.Postgres.DbString)
	if err != nil {
		return nil, err
	}
	dbConfig.MaxConns = 4
	dbConfig.MinConns = 0
	dbConfig.MaxConnLifetime = time.Hour
	dbConfig.MaxConnIdleTime = time.Minute * 30
	dbConfig.HealthCheckPeriod = time.Minute
	dbConfig.ConnConfig.ConnectTimeout = time.Second * 5
	return dbConfig, nil
}

func Initialize(ctx context.Context, conf *config.ConfigYaml, logger zerolog.Logger) (*DatabaseStr, error) {
	db := new(DatabaseStr)
	dbConfig, err := dbConfig(conf)
	if err != nil {
		return nil, err
	}
	connPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}
	db.PgxPool = connPool
	db.conf = conf
	db.logger = logger
	return db, err
}

func (d *DatabaseStr) Ping(ctx context.Context) error {
	err := d.PgxPool.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *DatabaseStr) SearchUserByEmail(ctx context.Context, email string) ([]Account, error) {
	rows, err := d.PgxPool.Query(ctx, `
		SELECT id, username, email, registration_date, edit_date, password_hashed, enabled, activated
		FROM gridpulse.accounts
		WHERE email=@email;
	`, pgx.NamedArgs{
		"email": email,
	})
	if err != nil {
		return nil, err
	}
	accounts, err := pgx.CollectRows(rows, pgx.RowToStructByName[Account])
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return accounts, nil
}

func (d *DatabaseStr) SearchUserByName(ctx context.Context, userName string) (*Account, error) {
	rows, err := d.PgxPool.Query(ctx, `
		SELECT id, username, email, registration_date, edit_date, password_hashed, enabled, activated
		FROM gridpulse.accounts
		WHERE username=@username;
	`, pgx.NamedArgs{
		"username": userName,
	})
	if err != nil {
		return nil, err
	}
	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Account])
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (d *DatabaseStr) AddUser(ctx context.Context, userName, passwordHash, email, authmethod, userrole string) error {
	_, err := d.PgxPool.Exec(ctx, `
		INSERT INTO gridpulse.accounts
		(username, email, registration_date, edit_date, password_hashed, auth_method, user_role)
		VALUES(@userName, @email, now(), now(), @passwordHashed, @authmethod, @userrole);
	`, pgx.NamedArgs{
		"userName":       userName,
		"email":          email,
		"passwordHashed": passwordHash,
		"authmethod":     authmethod,
		"userrole":       userrole,
	})
	if err != nil {
		return err
	}
	return nil
}
