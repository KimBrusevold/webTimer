package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/KimBrusevold/webTimer/timer"

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
var host string
var port string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	} else {
		log.Print("Loaded variables from .env file")
	}
	connStr, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
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
	r.HandleFunc("/registrer-bruker", RegisterHandler)
	r.HandleFunc("/auth/authenticate/{onetimeCode}", AuthenticateHandler)
	//MIDLERTIDIGE
	r.HandleFunc("/hent-bruker", GetHandler)

	port, exists = os.LookupEnv("PORT")
	if !exists {
		log.Fatal("No env variable named 'PORT' in .env file or environment variable. Exiting")
	}
	host, exists = os.LookupEnv("HOSTURL")
	if !exists {
		log.Fatal("No env variable named 'HOSTURL' in .env file or environment variable. Exiting")
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

func AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	onetimeCode := vars["onetimeCode"]

	if onetimeCode == "" {
		w.Header().Add("Location", "/registrer-bruker")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	user, err := timerDb.UserAuthProcees(onetimeCode)
	if err != nil {
		log.Fatal(err)
	}

	userAuthCookie := http.Cookie{
		Name:     "userAuthCookie",
		Value:    user.Authcode.String, //Should be some random authString. Signed? "github.com/go-http-utils/cookie"
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}

	userIdCookie := http.Cookie{
		Name:     "userId",
		Value:    strconv.FormatInt(user.ID, 10), //Should be some random authString. Signed? "github.com/go-http-utils/cookie"
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}

	http.SetCookie(w, &userAuthCookie)
	http.SetCookie(w, &userIdCookie)
	log.Print("Setting Cookie")
	w.Header().Add("Location", "/")
	w.WriteHeader(http.StatusSeeOther)
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

func SendAuthMail(userId int64, email string) error {
	res, err := timerDb.GetUser(userId)
	if err != nil {
		log.Print("Could not retrieve")
		log.Print(err.Error())
		return err
	}

	templateId := "d-67cb50f335c44f85a8960612dc97e7bc"
	httpposturl := "https://api.sendgrid.com/v3/mail/send"

	redirectUrl := fmt.Sprintf("%s/auth/authenticate/authenticate/%s", host, res.OneTimeCode.String)

	postString := fmt.Sprintf(`{
		"from":{
			"email":"kim.brusevold@soprasteria.com"
		 },
		 "personalizations":[
			{
			   "to":[
				  {
					 "email":"%s"
				  }
			   ],
			   "dynamic_template_data":{
				  "url": "%s"
				}
			}
		 ],
		 "template_id":"%s"
	}`, email, redirectUrl, templateId)

	log.Print(postString)

	postdata := []byte(postString)
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(postdata))
	if err != nil {
		log.Fatal("Klarte ikke å forberede request for å sende template")
	}
	request.Header.Set("Content-Type", "application/json;")
	sendgridApiKey, exists := os.LookupEnv("SENDGRID_API_KEY")
	if !exists {
		log.Fatal("No env variable named 'SENDGRID_API_KEY' in .env file or environment variable. Exiting")
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sendgridApiKey))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Panic(err)
	}
	defer response.Body.Close()

	log.Printf("Sendgripd response Status: %s", response.Status)

	return nil
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
		log.Print("User not authenticated")
		log.Print(cErr.Error())
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
