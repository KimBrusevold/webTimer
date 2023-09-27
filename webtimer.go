package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/KimBrusevold/webTimer/handlers"
	"github.com/KimBrusevold/webTimer/timer"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	// _, exists := os.LookupEnv("DATABASE_URL")
	// if !exists {
	// 	log.Fatal("No env variable named 'DB_CONN_STR' in .env file or environment variable. Exiting")
	// }

	// conn, err := sql.Open("pgx", connStr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer conn.Close()

	// log.Print("Migrating sqlLite")
	// timerDb = timer.NewDbTimerRepository(conn)
	// if err := timerDb.Migrate(); err != nil {
	// 	log.Fatal(err)
	// }

	// dir := "./static/images"

	r := gin.Default()
	r.LoadHTMLGlob("./pages/*")

	r.GET("/", HomeHandler)
	// r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir(dir))))
	r.Static("/images", "./static/images")
	// r.HandleFunc("/end", EndTimerHandler)
	registrerBruker := r.Group("/registrer-bruker")
	handlers.HandleRegisterUser(registrerBruker)

	// r.HandleFunc("/auth/authenticate/{onetimeCode}", AuthenticateHandler)
	// //MIDLERTIDIGE
	// r.HandleFunc("/hent-bruker", GetHandler)

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
		Secure:   true,
		Path:     "/",
	}

	userIdCookie := http.Cookie{
		Name:     "userId",
		Value:    strconv.FormatInt(user.ID, 10), //Should be some random authString. Signed? "github.com/go-http-utils/cookie"
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}

	http.SetCookie(w, &userAuthCookie)
	http.SetCookie(w, &userIdCookie)

	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte("You are ready to time go to /"))
}

func EndTimerHandler(w http.ResponseWriter, r *http.Request) {
	userCookie, cErr := r.Cookie("userAuthCookie")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userAuthCookie")
		log.Print(cErr.Error())
		w.Header().Add("Location", "/registrer-bruker")
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	idCookie, cErr := r.Cookie("userId")

	if cErr != nil {
		log.Print("User not authenticated. Does not have userId cookie")
		log.Print(cErr.Error())
		w.Header().Add("Location", "/registrer-bruker")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	i, err := strconv.Atoi(idCookie.Value)
	if err != nil {
		log.Print("Could not get id from cookie")
		log.Print(cErr.Error())
		w.Header().Add("Location", "/registrer-bruker")
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	isAuthenticated := timerDb.IsAuthorizedUser(userCookie.Value, i)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		log.Print(cErr.Error())
		w.Header().Add("Location", "/registrer-bruker")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	timeUsed, err := timerDb.EndTimeTimer(i)
	if err != nil {
		log.Print("Could not stop timer")
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	minutes := timeUsed / (60 * 1000) % 60
	seconds := timeUsed / (1000) % 60
	tenths := timeUsed / (100) % 1000
	_, e := w.Write([]byte(fmt.Sprintf("<h1>Du brukte %d min %d.%d sekunder<h2>", minutes, seconds, tenths)))
	if e != nil {
		log.Fatal(e)
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

func HomeHandler(c *gin.Context) {
	userCookie, cErr := c.Cookie("userAuthCookie")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userAuthCookie")
		log.Print(cErr.Error())
		c.Header("Location", "/registrer-bruker")
		c.Status(http.StatusSeeOther)
		return
	}
	idCookie, cErr := c.Cookie("userId")

	if cErr != nil {
		log.Print("User not authenticated. Does not have userId cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/registrer-bruker")
		c.Status(http.StatusSeeOther)
		return
	}

	i, err := strconv.Atoi(idCookie)
	if err != nil {
		log.Print("Could not get id from cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/registrer-bruker")
		c.Status(http.StatusSeeOther)
		return
	}
	isAuthenticated := timerDb.IsAuthorizedUser(userCookie, i)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		log.Print(cErr.Error())
		c.Header("Location", "/registrer-bruker")
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
