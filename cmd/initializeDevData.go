package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/joho/godotenv"
)

type settings struct {
	hostUrl            string
	dbUrl              string
	tursoAuthToken     string
	port               string
	senderEmailAddress string
	emailPassword      string
}

func main() {
	settings := getEnvSettings()

	connStr := buildConnectionString(settings)

	db, err := sql.Open("libsql", connStr)
	pingErr := db.Ping()
	log.Printf("Opening and pinging %s", connStr)
	if err != nil {
		panic(pingErr)
	}
	if err != nil {
		log.Fatalf("Could not create connector to database: %s", err)
	}
	defer db.Close()

	timerDb := database.NewDbTimerRepository(db)

	createMockUsers(timerDb)
	createMockTimes(db)
}

func createMockTimes(timerDb *sql.DB) {
	location := time.Now().Location()
	startTime := time.Date(2024, 03, 1, 0, 0, 0, 0, location).UTC()

	for i := 1; i <= 3; i++ {
		userid := i + 1
		for i := 0; i < 10; i++ {
			timerDb.Exec(`INSERT INTO times()`)
		}
	}
}

func createMockUsers(db *database.TimerDB) {
	db.CreateUser(database.User{
		Username: "testuser1",
		Email:    "test@email.com",
		Password: "1234",
	})

	db.CreateUser(database.User{
		Username: "trappesønn",
		Email:    "trapp@gmail.com",
		Password: "1234",
	})
	db.CreateUser(database.User{
		Username: "sjefen",
		Email:    "serius@business.com",
		Password: "1234",
	})
}

func getEnvSettings() settings {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	} else {
		log.Print("Loaded variables from .env file")
	}

	hostUrl, exists := os.LookupEnv("HOSTURL")
	if !exists {
		log.Fatal("No rnv variable named 'HOSTURL' in .env file or environment variable. Exiting")
	}

	dbUrl, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatal("No env variable named 'DATABASE_URL' in .env file or environment variable. Exiting")
	}
	authToken, exists := os.LookupEnv("TURSO_AUTH_TOKEN")
	if !exists {
		log.Fatal("No env variable named 'TURSO_AUTH_TOKEN' in .env file or environment variable. Exiting")
	}

	senderEmail, exists := os.LookupEnv("EMAIL_SENDER_ADDRESS")
	if !exists || senderEmail == "" {
		log.Fatal("No env variable or emtpy value named 'EMAIL_SENDER_ADDRESS' in .env file or environment variable. Exiting")
	}

	emailPassword, exists := os.LookupEnv("EMAIL_PASSWORD")
	if !exists || emailPassword == "" {
		log.Fatal("No env variable or emtpy value named 'EMAIL_PASSWORD' in .env file or environment variable. Exiting")
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		log.Println("No port set. Using default: 8080")
		port = "8080"
	}

	return settings{
		hostUrl:            hostUrl,
		dbUrl:              dbUrl,
		tursoAuthToken:     authToken,
		port:               port,
		senderEmailAddress: senderEmail,
		emailPassword:      emailPassword,
	}
}

func buildConnectionString(s settings) string {
	var connString string
	if s.tursoAuthToken == "" {
		connString = s.dbUrl
	} else {
		connString = s.dbUrl + fmt.Sprintf("?authToken=%s", s.tursoAuthToken)
	}
	return connString
}
