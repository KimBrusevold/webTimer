package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/KimBrusevold/webTimer/internal/handlers"
	"github.com/KimBrusevold/webTimer/internal/timer"
	"github.com/joho/godotenv"

	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

var timerDb *timer.TimerDB
var host string
var port string

func main() {
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
	connStr, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatal("No env variable named 'DATABASE_URL' in .env file or environment variable. Exiting")
	}
	log.Print("DATABASE_URL is provided")
	conn, err := sql.Open("libsql", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Print("Migrating sqlLite")
	timerDb = timer.NewDbTimerRepository(conn)
	if err := timerDb.Migrate(); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("./web/pages/template/**/*")

	r.GET("/", leaderboard)

	r.Use(func(c *gin.Context) {
		// Apply the Cache-Control header to the script static file
		if strings.HasPrefix(c.Request.URL.Path, "/res/scripts") {
			c.Header("Cache-Control", "max-age=31536000, immutable")
		}
		// Continue to the next middleware or handler
		c.Next()
	})

	r.Static("/res/images", "./web/static/images")
	r.Static("/res/css", "./web/static/css")
	r.Static("/res/scripts", "./web/static/scripts")

	r.StaticFile("/favicon.ico", "./web/static/images/upstairs.png")

	registrerBruker := r.Group("/registrer-bruker")
	authGroup := r.Group("/autentisering")

	r.Use(authenticate)
	r.GET("/timer/start-lop", startTimerHandler)
	r.GET("/timer/avslutt-lop", endTimerHandler)

	handlers.HandleRegisterUser(registrerBruker, timerDb, host)
	handlers.HandleAuthentication(authGroup)

	port, exists := os.LookupEnv("PORT")
	if !exists {
		log.Fatal("No env variable named 'PORT' in .env file or environment variable. Exiting")
	}
	// host, exists = os.LookupEnv("HOSTURL")
	// if !exists {
	// 	log.Fatal("No env variable named 'HOSTURL' in .env file or environment variable. Exiting")
	// }

	addr := fmt.Sprintf("0.0.0.0:%s", port)

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Now listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

type TimesDisplay struct {
	Place    int
	Username string
	Minutes  int64
	Seconds  int64
	Tenths   int64
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
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	isAuthenticated := timerDb.IsAuthorizedUser(userCookie, i)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		log.Print(cErr.Error())
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
		// minutes :=  / (60 * 1000) % 60
		// seconds := timeUsed / (1000) % 60
		// tenths := timeUsed / (100) % 1000
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

	if err != nil {
		log.Print("ERROR: Could not read tid-startet.html from file")
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "tid-startet.tmpl", nil)

}
