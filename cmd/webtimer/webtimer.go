package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/KimBrusevold/webTimer/internal/email"
	"github.com/KimBrusevold/webTimer/internal/handler"
	"github.com/KimBrusevold/webTimer/internal/handler/auth"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite" // this dependency for running libsql from a db file
)

var timerDb *database.TimerDB

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

	timerDb = database.NewDbTimerRepository(db)

	r := gin.Default()
	r.LoadHTMLGlob("./web/pages/template/**/*")

	lh := handler.LeaderboardHandler{
		DB: timerDb,
	}
	
	r.GET("/", lh.HandleLeaderboardShow)
	r.GET("/leaderboard/raskest", lh.RenderFastestLeaderboard)

	r.Static("/res/images", "./web/static/images")
	r.Static("/res/css", "./web/static/css")
	r.Static("/res/scripts", "./web/static/scripts")

	r.StaticFile("/favicon.ico", "./web/static/images/upstairs.png")

	authHandler := auth.AuthHandler{
		DB: timerDb,
		EmailClient: &email.EmailClient{
			HostAddr:   "smtp.gmail.com",
			SenderAddr: settings.senderEmailAddress, // A gmail address
			Password:   settings.emailPassword,      // A gmail app key
		},
	}
	authHandler.SetupRoutes(r.Group("/aut"))

	timerH := handler.TimerHandler{
		DB: timerDb,
	}
	timerH.SetupRoutes(r.Group("/timer"))

	addr := fmt.Sprintf("0.0.0.0:%s", settings.port)

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Now listening on %s", addr)

	err = srv.ListenAndServe()

	log.Printf("Shutting down server %s", err.Error())
	os.Exit(0)
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
