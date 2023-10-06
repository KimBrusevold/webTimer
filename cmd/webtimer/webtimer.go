package main

import (
	"database/sql"
	"fmt"
	"strconv"

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

	// dir := "./static/images"

	r := gin.Default()
	r.LoadHTMLGlob("./web/pages/**/*")
	r.GET("/", HomeHandler)
	r.GET("/avslutt-lop", EndTimerHandler)
	r.GET("/leaderboard", leaderboard)
	r.Static("/images", "./web/static/images")
	r.Static("/css", "./web/static/css")
	r.StaticFile("/favicon.ico", "./web/static/images/upstairs.png")
	registrerBruker := r.Group("/registrer-bruker")
	authGroup := r.Group("/autentisering")
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
	Place   int
	UserID  int64
	Minutes int64
	Seconds int64
	Tenths  int64
}

func leaderboard(c *gin.Context) {
	times, err := timerDb.RetrieveTimes()
	if err != nil {
		log.Printf("Could not get times from db. %s", err.Error())
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	var timesDisplay []TimesDisplay

	for i, t := range times {
		// minutes :=  / (60 * 1000) % 60
		// seconds := timeUsed / (1000) % 60
		// tenths := timeUsed / (100) % 1000
		td := TimesDisplay{
			Place:   i + 1,
			UserID:  t.UserID,
			Minutes: t.ComputedTime.Int64 / (60 * 1000) % 60,
			Seconds: t.ComputedTime.Int64 / (1000) % 60,
			Tenths:  t.ComputedTime.Int64 / (100) % 1000,
		}
		timesDisplay = append(timesDisplay, td)
	}

	c.HTML(http.StatusOK, "leaderboard.tmpl", timesDisplay)
}

func EndTimerHandler(c *gin.Context) {
	userCookie, cErr := c.Cookie("userAuthCookie")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userAuthCookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}
	idCookie, cErr := c.Cookie("userId")

	if cErr != nil {
		log.Print("User not authenticated. Does not have userId cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}

	userId, err := strconv.Atoi(idCookie)
	if err != nil {
		log.Print("Could not get id from cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}
	isAuthenticated := timerDb.IsAuthorizedUser(userCookie, userId)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}

	timeUsed, err := timerDb.EndTimeTimer(userId)
	if err != nil {
		log.Print("Could not stop timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	minutes := timeUsed / (60 * 1000) % 60
	seconds := timeUsed / (1000) % 60
	tenths := timeUsed / (100) % 1000
	c.String(http.StatusOK, "<h1>Du brukte %d min %d.%d sekunder<h2>", minutes, seconds, tenths)

}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	res, err := timerDb.GetUser(1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not retrieve")
		log.Print(err.Error())
		return
	}

	w.Write([]byte(res.Username))

}

func HomeHandler(c *gin.Context) {
	userCookie, cErr := c.Cookie("userAuthCookie")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userAuthCookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}
	idCookie, cErr := c.Cookie("userId")

	if cErr != nil {
		log.Print("User not authenticated. Does not have userId cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}

	i, err := strconv.Atoi(idCookie)
	if err != nil {
		log.Print("Could not get id from cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}
	isAuthenticated := timerDb.IsAuthorizedUser(userCookie, i)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		log.Print(cErr.Error())
		c.Header("Location", "/autentisering/login")
		c.Status(http.StatusSeeOther)
		return
	}

	err = timerDb.StartTimer(i)
	if err != nil {
		log.Print("Could not start timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if err != nil {
		log.Print("ERROR: Could not read index.html from file")
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "index.html", nil)

}

func IsAuthorizedUser(s string, i int) {
	panic("unimplemented")
}
