package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file detected. Will pretend nothing happened")
		err = nil
	}

	if os.Getenv("POSTGRES_HOST") == "" {
		log.Fatalln("No `POSTGRES_HOST` env variable detected. Didn't you forget to load the .env (locally) or inject env variables onto docker/kuber?")
	}

	m, err := migrate.New(
		"file://db/migrations",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_SSLMODE")),
	)

	if err != nil {
		log.Fatal(err)
	}

	version, dirty, _ := m.Version()

	log.Printf("Current version: %d, dirty?: %t\n", version, dirty)

	if len(os.Args) <= 1 {
		fmt.Println()
		log.Fatalln("In order to run a migration you need to say what migration method to use. The options are: up or down")
	}

	mode := os.Args[1]

	switch mode {
	case "up":
		err = m.Up()
		break
	case "down":
		err = m.Down()
		break
	default:
		break
	}

	if err != nil {
		if err.Error() == "no change" {
			log.Println("Database up to date")
		} else {
			log.Fatalf("Error while executing migration: `%v`", err)
		}
	}

	log.Print("All done\n")

	version, dirty, _ = m.Version()

	log.Printf("New version: %d, dirty?: %t\n", version, dirty)
}
