package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	boardAPI "server/api/board"
	loginAPI "server/api/login"
	registerAPI "server/api/register"
	"server/cookie"
	"server/db"
)

func main() {
	// Create dependencies that are shared by multiple handlers.
	conn, err := sql.Open("postgres", os.Getenv("DBCONNSTR"))
	if err != nil {
		log.Fatal(err)
	}
	connCloser := db.NewConnCloser(conn)
	jwtGenerator := cookie.NewJWTGenerator(os.Getenv("JWTSIGNATURE"))
	userReader := db.NewUserReader(conn)

	// Register handlers for API endpoints.
	mux := http.NewServeMux()
	mux.Handle("/register", registerAPI.NewHandler(
		registerAPI.NewRequestValidator(
			registerAPI.NewUsernameValidator(),
			registerAPI.NewPasswordValidator(),
		),
		userReader,
		registerAPI.NewPasswordHasher(),
		db.NewUserCreator(conn),
		jwtGenerator,
		connCloser,
	))
	mux.Handle("/login", loginAPI.NewHandler(
		userReader,
		loginAPI.NewPasswordComparer(),
		jwtGenerator,
		connCloser,
	))
	mux.Handle("/board", boardAPI.NewHandler())

	// Serve the app using the ServeMux.
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
