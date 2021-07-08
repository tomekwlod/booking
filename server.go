package main

// https://stackoverflow.com/questions/63734176/how-to-wrap-sql-transaction-in-go-with-existing-repository-that-doesnt-use-sql
// https://github.com/drone/drone/blob/master/store/user/user.go

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
	"github.com/tomekwlod/booking/internal/database"
	"github.com/tomekwlod/booking/model"
	userRepo "github.com/tomekwlod/booking/repo/user"
	"github.com/tomekwlod/utils/env"
)

var pconn *database.PConn

func main() {

	err := godotenv.Load()

	if err != nil {

		log.Println("No .env file detected. Will pretend nothing happened")
		err = nil
	}

	if os.Getenv("PROJECT_NAME") == "" {

		log.Fatalln("No `PROJECT_NAME` env variable detected. Didn't you forget to load the .env (locally) or inject env variables onto docker/kuber?")
	}

	// jwtSecret = env.Env("JWT_SECRET", "")

	// if jwtSecret == "" {

	// 	return errors.New("env: JWT_SECRET not set but needed for the server to run")
	// }

	// jwtCookieName = env.Env("JWT_COOKIE_NAME", "jwt_cookie")

	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env.Env("POSTGRES_HOST", "localhost"),
		env.Env("POSTGRES_PORT", "5432"),
		env.Env("POSTGRES_USER", "postgres"),
		env.Env("POSTGRES_PASSWORD", "postgres"),
		env.Env("POSTGRES_DB", "booking"),
		env.Env("POSTGRES_SSLMODE", "disable"))

	pconn, err = database.PostgresConnection(dbURL)

	if err != nil {

		log.Fatalf("error while connecting to db %v", err)
	}

	defer pconn.Close()

	tx, _ := pconn.Begin()
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			err = tx.Rollback() // or use a []error, or else, to not shadow the underlying error.
		}
	}()
	ctx := context.Background()

	user := &model.User{
		Username:    "twl",
		Email:       "twl",
		Description: "testonly",
	}

	fmt.Printf("%+v\n", user)

	var ur userRepo.UserRepo
	err = ur.Create(ctx, tx, user)
	if err != nil {
		panic(err)
	}

	// fmt.Println(id)
	fmt.Printf("%+v\n", user)

	return

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	regularMiddlewares := []Middleware{
		recoverMiddleware(),

		// // the logging & tracing middlewares were moved to the server declaration to wrap all the traffic
		// loggingMiddleware,
		// tracing(nextRequestID),
	}

	router := mux.NewRouter()

	router.Handle("/", Chain(
		http.HandlerFunc(indexHandler),

		append(
			regularMiddlewares,
			// withJWT,
			// withHeaderMiddleware("X-Something2", "Specific"),
			supportXHTTPMethodOverrideMiddleware(),
		)...,
	)).Methods("GET")

	router.Handle("/dbtest", Chain(
		http.HandlerFunc(dbtestHandler),

		append(
			regularMiddlewares,
			supportXHTTPMethodOverrideMiddleware(),
		)...,
	)).Methods("GET")

	// router.Handle("/login", Chain(
	// 	http.HandlerFunc(loginHandler),

	// 	append(
	// 		regularMiddlewares,
	// 		// withHeaderMiddleware("X-Something2", "Specific"),
	// 		supportXHTTPMethodOverrideMiddleware(),
	// 	)...,
	// )).Methods("POST")

	// router.Handle("/logout", Chain(
	// 	http.HandlerFunc(logoutHandler),

	// 	append(
	// 		regularMiddlewares,
	// 		supportXHTTPMethodOverrideMiddleware(),
	// 	)...,
	// )).Methods("GET")

	// router.Handle("/signup", Chain(
	// 	http.HandlerFunc(signupHandler),

	// 	append(
	// 		regularMiddlewares,
	// 		// withHeaderMiddleware("X-Something2", "Specific"),
	// 		supportXHTTPMethodOverrideMiddleware(),
	// 	)...,
	// )).Methods("POST")

	// router.Handle("/welcome", Chain(
	// 	http.HandlerFunc(welcomeHandler),

	// 	append(
	// 		regularMiddlewares,
	// 		withJWT,
	// 		// withHeaderMiddleware("X-Something2", "Specific"),
	// 		supportXHTTPMethodOverrideMiddleware(),
	// 	)...,
	// )).Methods("GET")

	router.Handle("/json", Chain(
		http.HandlerFunc(jsonHandler),

		append(
			regularMiddlewares,
			// withHeaderMiddleware("X-Something2", "Specific"),
			supportXHTTPMethodOverrideMiddleware(),
		)...,
	)).Methods("GET")

	router.Handle("/upload", Chain(
		http.HandlerFunc(uploadHandler),

		append(
			regularMiddlewares,
			supportXHTTPMethodOverrideMiddleware(),
		)...,
	)).Methods("POST")

	host := "127.0.0.1"
	port := env.Env("API_SRV_PORT", "8080")

	srv := &http.Server{
		Handler: Chain(
			router,
			// all routes needs logging and tracing
			loggingMiddleware, tracing(nextRequestID),
		),
		Addr: fmt.Sprintf("%s:%s", host, port),
		// Good practice: enforce timeouts for servers we create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {

		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {

			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("Server started and listening on port %s.\n\n", port)

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Waiting for SIGINT (pkill -2) eg.: ctrl+c
	<-stop

	// Server will shutdown immediately unless there are still some open connections
	// It is still important to close the server after some amount of time for example when
	// some zombies are constantly using our server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// Close database, redis, truncate message queues, etc. here
		cancel()
	}()

	err = srv.Shutdown(ctx)

	if err != nil {

		log.Fatalf("Server shutdown failed: %s\n", err.Error())
	}

	log.Println("Gracefully server shutdown")
}
