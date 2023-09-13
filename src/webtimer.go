package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"kimnb/webtimer/timer"

	"encoding/json"
	"io"
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
	//MIDLERTIDIGE
	r.HandleFunc("/hent-bruker", GetHandler)

	port, exists = os.LookupEnv("PORT")
	if exists == false {
		log.Fatal("No env variable named 'PORT' in .env file or environment variable. Exiting")
	}
	host, exists = os.LookupEnv("HOSTURL")
	if exists == false {
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
			log.Printf("Error on create: \n")
			log.Printf(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("%d", userid)))

		_ = SendAuthMail(userid, email)
		return
	}

}
func GetHandler(w http.ResponseWriter, r *http.Request) {
	res, err := timerDb.GetUser(1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not retrieve")
		log.Printf(err.Error())
		return
	}

	w.Write([]byte(fmt.Sprintf("%s", res.Username)))

}

func SendAuthMail(userId int64, email string) error {
	res, err := timerDb.GetUser(userId)
	if err != nil {
		log.Printf("Could not retrieve")
		log.Printf(err.Error())
		return err
	}

	templateId := "d-67cb50f335c44f85a8960612dc97e7bc"
	httpposturl := "https://api.sendgrid.com/v3/mail/send"

	redirectUrl := fmt.Sprintf("%s:%s/authenticate/%s", host, port, res.OneTimeCode.String)

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

	log.Printf(postString)

	postdata := []byte(postString)
	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(postdata))
	if error != nil {
		log.Fatal("Klarte ikke å forberede request for å sende template")
	}
	request.Header.Set("Content-Type", "application/json;")
	sendgridApiKey, exists := os.LookupEnv("SENDGRID_API_KEY")
	if exists == false {
		log.Fatal("No env variable named 'SENDGRID_API_KEY' in .env file or environment variable. Exiting")
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sendgridApiKey))

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		log.Panic(error)
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	log.Printf("response Body :%s", string(body))

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
