package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(
	ctx context.Context,
	host string,
	port int,
	user string,
	password string,
	dbname string,
) (*pgxpool.Pool, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		user,
		password,
		host,
		port,
		dbname,
	)

	return pgxpool.New(ctx, dsn)
}
