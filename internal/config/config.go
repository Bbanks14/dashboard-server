package config

import "os"

var (
	DBURL     = os.Getenv("postgres://tayo:passw0rd@localhost:5432/dashboard_server?sslmode=disable")
	JWTSecret = os.Getenv("MyInititalJWTSecret")
)
