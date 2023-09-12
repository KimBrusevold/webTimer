package main

import (
	"database/sql"
	"fmt"
	"kimnb/webtimer/timer"

	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var timerDb *timer.TimerDB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	} else {
		log.Print("Loaded variables from .env file")
	}
	connStr, exists := os.LookupEnv("DB_CONN_STR")
	if exists == false {
		log.Fatal("No env variable named 'DB_CONN_STR' in .env file or environment variable. Exiting")
	}

	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Print("Migrating sqlLite")
	timerDb = timer.NewDbTimerRepository(conn)
	if err := timerDb.Migrate(); err != nil {
		log.Fatal(err)
	}

	dir := "./static/images"

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir(dir))))
	r.HandleFunc("/start", StartTimerHandler)
	r.HandleFunc("/end", EndTimerHandler)
	r.HandleFunc("/setCookie", SetCookieHandler)
	r.HandleFunc("/registrer-bruker", RegisterHandler)

	port, exists := os.LookupEnv("PORT")
	if exists == false {
		log.Fatal("No env variable named 'PORT' in .env file or environment variable. Exiting")
	}

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

func SetCookieHandler(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:     "userAuthCookie",
		Value:    "123", //Should be some random authString. Signed? "github.com/go-http-utils/cookie"
		Expires:  time.Now().Add(15 * time.Second),
		HttpOnly: true,
	}
	http.SetCookie(w, &c)

	w.WriteHeader(http.StatusOK)
}

func StartTimerHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().UnixMilli()
	t := timer.Timer{
		StartTime: startTime,
		EndTime:   0,
	}
	log.Print("Creating TIme")
	createTime, err := timerDb.Create(t)

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Print("Creating JSON")
	jTime, err := json.Marshal(createTime)

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Print("WRITING JSON")

	w.Write(jTime)
}

func EndTimerHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%#v", time.Now().UnixMilli())
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		content, err := os.ReadFile("./pages/register-user.html")
		if err != nil {
			log.Print("ERROR: Could not read register-user.html from file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, e := w.Write(content)
		if e != nil {
			log.Fatal(e)
			return
		}
	} else if r.Method == "POST" {
		r.ParseForm()

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// timerDb.Create()
		log.Printf("Username: %s \n", username)
		log.Printf("Email: %s \n", email)
		log.Printf("Password: %s \n", password)
	}

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("./pages/index.html")
	if err != nil {
		log.Print("ERROR: Could not read index.html from file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, cErr := r.Cookie("userAuthCookie")
	if cErr != nil {
		w.Header().Add("Location", "/registrer-bruker")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	print("Should validate cookie")

	_, e := w.Write(content)
	if e != nil {
		log.Fatal(e)
		return
	}
}
