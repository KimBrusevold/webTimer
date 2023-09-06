package main

import (
	"database/sql"
	"encoding/json"
	"kimnb/webtimer/timer"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const fileName = "sqlite.db"

var timerDb *timer.TimerDB

func main() {
	os.Remove(fileName)
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Migrating sqlLite")
	timerDb = timer.NewSQLiteRepository(db)
	if err := timerDb.Migrate(); err != nil {
		log.Fatal(err)
	}

	dir := "./static/images"

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir(dir))))
	r.HandleFunc("/start", StartTimerHandler)
	r.HandleFunc("/end", EndTimerHandler)
	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
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

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("./pages/index.html")

	if err != nil {
		log.Fatal(err)
	}

	_, e := w.Write(content)
	if e != nil {
		log.Fatal(e)
		return
	}

	w.WriteHeader(http.StatusOK)
}
