package config

import "fmt"

func BuildDSN(
	host string,
	port int,
	user string,
	password string,
	dbname string,
) string {

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname,
	)
}
