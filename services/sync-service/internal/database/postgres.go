package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open abre la conexión PostgreSQL utilizada por Sync Service.
//
// Se utilizan las mismas variables de entorno y valores por
// defecto que File Service para evitar configuraciones distintas
// dentro de la misma plataforma.
func Open() (*sql.DB, error) {

	host := getenv(
		"POSTGRES_HOST",
		"localhost",
	)

	port := getenv(
		"POSTGRES_PORT",
		"5434",
	)

	user := getenv(
		"POSTGRES_USER",
		"upb_app",
	)

	password :=
		os.Getenv(
			"SYNC_DB_PASSWORD",
		)

	if password == "" {
		password =
			os.Getenv(
				"POSTGRES_PASSWORD",
			)
	}

	if password == "" {
		return nil,
			fmt.Errorf(
				"SYNC_DB_PASSWORD o POSTGRES_PASSWORD es obligatorio",
			)
	}

	databaseName := getenv(
		"POSTGRES_DB",
		"upb_cientifica",
	)

	dsn :=
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			host,
			port,
			databaseName,
		)

	db, err :=
		sql.Open(
			"pgx",
			dsn,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo abrir PostgreSQL: %w",
				err,
			)
	}

	db.SetMaxOpenConns(
		10,
	)

	db.SetMaxIdleConns(
		5,
	)

	db.SetConnMaxLifetime(
		30 * time.Minute,
	)

	if err :=
		db.Ping(); err != nil {

		_ =
			db.Close()

		return nil,
			fmt.Errorf(
				"no se pudo conectar con PostgreSQL: %w",
				err,
			)
	}

	return db, nil
}

func getenv(
	key string,
	fallback string,
) string {

	value :=
		os.Getenv(
			key,
		)

	if value == "" {
		return fallback
	}

	return value
}
