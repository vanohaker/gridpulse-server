package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInit, downInit)
}

func upInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS gridpulse;
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE TABLE gridpulse.accounts (
			id uuid DEFAULT uuid_generate_v4() NOT NULL, -- User UUID
			username varchar NOT NULL, -- Username
			email varchar NOT NULL, -- User's email address
			registration_date timestamptz NOT NULL, -- User registration date
			edit_date timestamptz NOT NULL, -- Profile Modification Date
			password_hashed varchar NULL, -- Password hash
			enabled bool NULL, -- Indication that the user is active
			activated bool NULL, -- Indication that the user is activated
			auth_method varchar NOT NULL, -- Authorization method
			user_role varchar NOT NULL, -- User role
			CONSTRAINT accounts_pk PRIMARY KEY (id),
			CONSTRAINT accounts_unique UNIQUE (username),
			CONSTRAINT accounts_unique_1 UNIQUE (email)
		);

		COMMENT ON COLUMN gridpulse.accounts.id IS 'User UUID';
		COMMENT ON COLUMN gridpulse.accounts.username IS 'Username';
		COMMENT ON COLUMN gridpulse.accounts.email IS 'User''s email address';
		COMMENT ON COLUMN gridpulse.accounts.registration_date IS 'User registration date';
		COMMENT ON COLUMN gridpulse.accounts.edit_date IS 'Profile Modification Date';
		COMMENT ON COLUMN gridpulse.accounts.password_hashed IS 'Password hash';
		COMMENT ON COLUMN gridpulse.accounts.enabled IS 'Indication that the user is active';
		COMMENT ON COLUMN gridpulse.accounts.activated IS 'Indication that the user is activated';
		COMMENT ON COLUMN gridpulse.accounts.auth_method IS 'Authorization method';
		COMMENT ON COLUMN gridpulse.accounts.user_role IS 'User role';
	`)
	if err != nil {
		return err
	}
	return nil
}

func downInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS gridpulse.accounts;
		DROP SCHEMA IF EXISTS gridpulse;
	`)
	if err != nil {
		return err
	}
	return nil
}
