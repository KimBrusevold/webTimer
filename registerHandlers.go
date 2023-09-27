package main

import (
	"log"
	"net/http"
	"os"

	"github.com/KimBrusevold/webTimer/timer"
)

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
			log.Print("Could not write content to responseWriter")
			log.Print(e.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		r.ParseForm()

		username := r.FormValue("username")
		email := r.FormValue("email")

		// timerDb.Create()
		log.Printf("Username: %s \n", username)
		log.Printf("Email: %s \n", email)

		user := timer.User{
			Username: username,
			Email:    email,
		}

		userid, err := timerDb.CreateUser(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Print("Error on create: \n")
			log.Print(err.Error())
			return
		}

		err = SendAuthMail(userid, email)
		if err != nil {
			w.Write([]byte("Something went wrong"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		content, err := os.ReadFile("./pages/email-sent.html")
		if err != nil {
			log.Print("ERROR: Could not read email-sent.html from file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(content)
		if err != nil {
			log.Fatal(err)
			return
		}
		return
	}

}
