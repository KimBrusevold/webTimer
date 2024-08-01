package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/KimBrusevold/webTimer/internal/handlers"
	"github.com/KimBrusevold/webTimer/internal/timer"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var timerDb *timer.TimerDB
var host string

type settings struct {
	HostUrl        string
	DbUrl          string
	TursoAuthToken string
	Port           string
}

func main() {
	settings := getEnvSettings()

	connStr := settings.DbUrl + fmt.Sprintf("?authToken=%s", settings.TursoAuthToken)

	db, err := sql.Open("libsql", connStr)
	if err != nil {
		log.Fatalf("Could not create connector to database: %s", err)
	}
	defer db.Close()

	timerDb = timer.NewDbTimerRepository(db)

	r := gin.Default()
	r.LoadHTMLGlob("./web/pages/template/**/*")

	r.GET("/", leaderboard)

	r.Static("/res/images", "./web/static/images")
	r.Static("/res/css", "./web/static/css")
	r.Static("/res/scripts", "./web/static/scripts")

	r.StaticFile("/favicon.ico", "./web/static/images/upstairs.png")

	handlers.HandleRegisterUser(r.Group("/registrer-bruker"), timerDb, host)
	handlers.HandleAuthentication(r.Group("/autentisering"))

	r.Use(authenticate)
	r.GET("/timer/start-lop", startTimerHandler)
	r.GET("/timer/avslutt-lop", endTimerHandler)

	// host, exists = os.LookupEnv("HOSTURL")
	// if !exists {
	// 	log.Fatal("No env variable named 'HOSTURL' in .env file or environment variable. Exiting")
	// }

	addr := fmt.Sprintf("0.0.0.0:%s", settings.Port)

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

type TimesDisplay struct {
	Place    int
	Username string
	Minutes  int64
	Seconds  int64
	Tenths   int64
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
	host = hostUrl
	dbUrl, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatal("No env variable named 'DATABASE_URL' in .env file or environment variable. Exiting")
	}
	authToken, exists := os.LookupEnv("TURSO_AUTH_TOKEN")
	if !exists {
		log.Fatal("No env variable named 'TURSO_AUTH_TOKEN' in .env file or environment variable. Exiting")
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		log.Fatal("No env variable named 'PORT' in .env file or environment variable. Exiting")
	}

	return settings{
		HostUrl:        hostUrl,
		DbUrl:          dbUrl,
		TursoAuthToken: authToken,
		Port:           port,
	}
}

func authenticate(c *gin.Context) {
	userCookie, cErr := c.Cookie("userAuthCookie")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userAuthCookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	idCookie, cErr := c.Cookie("userId")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userId cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	i, err := strconv.Atoi(idCookie)
	if err != nil {
		log.Print("Could not get id from cookie")
		log.Print(err.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	isAuthenticated := timerDb.IsAuthorizedUser(userCookie, i)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	c.Set("userId", i)
}

func leaderboard(c *gin.Context) {
	times, err := timerDb.RetrieveTimes()
	if err != nil {
		log.Printf("Could not get times from db. %s", err.Error())
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	var timesDisplay []TimesDisplay

	for _, t := range times {
		td := TimesDisplay{
			Place:    t.Place,
			Username: t.Username,
			Minutes:  t.ComputedTime / (60 * 1000) % 60,
			Seconds:  t.ComputedTime / (1000) % 60,
			Tenths:   t.ComputedTime / (100) % 1000,
		}
		timesDisplay = append(timesDisplay, td)
	}

	c.HTML(http.StatusOK, "leaderboard.tmpl", gin.H{
		"title": "Resultatliste",
		"data":  timesDisplay,
	})

}

func endTimerHandler(c *gin.Context) {
	i, exists := c.Get("userId")
	if !exists {
		log.Print("Found no userId in context. Cannot start timer")
		c.Status(http.StatusInternalServerError)
		return
	}

	timeUsed, err := timerDb.EndTimeTimer(i.(int))
	if err != nil {
		log.Print("Could not stop timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	minutes := timeUsed / (60 * 1000) % 60
	seconds := timeUsed / (1000) % 60
	tenths := timeUsed / (100) % 1000
	c.HTML(http.StatusOK, "tid-avsluttet.tmpl", gin.H{
		"minutes": minutes,
		"seconds": seconds,
		"tenths":  tenths,
	})

}

func startTimerHandler(c *gin.Context) {
	i, exists := c.Get("userId")
	if !exists {
		log.Print("Found no userId in context. Cannot start timer")
		c.Status(http.StatusInternalServerError)
		return
	}
	err := timerDb.StartTimer(i.(int))
	if err != nil {
		log.Print("Could not start timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "tid-startet.tmpl", nil)
}
