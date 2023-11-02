package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/KimBrusevold/webTimer/internal/timer"
	"github.com/gin-gonic/gin"
)

var timerDb *timer.TimerDB
var hostUrl string

func HandleRegisterUser(rg *gin.RouterGroup, db *timer.TimerDB, host string) {
	rg.GET("/", registerUserPage)
	rg.POST("/", createUser)
	rg.POST("/email", validateEmail)
	rg.POST("/username", validateUsername)

	timerDb = db
	hostUrl = host
}

func registerUserPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register-user.html", nil)
}

func createUser(c *gin.Context) {
	user := timer.User{
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
	}
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)

	log.Printf("verifiserer og skaper bruker med brukernavn: %s, og epost: %s", user.Username, user.Email)
	//TODO: Denne bør sette error i form = Epost og brukernavn er påkrevde felter.
	if user.Email == "" || user.Username == "" {
		log.Print("Could not bind form data to user")
		c.String(http.StatusBadRequest, "Ugyldig epost eller navn")
		return
	}
	v := strings.Split(user.Email, "@")

	if len(v) != 2 {
		log.Printf("Invalid email: %s", user.Email)
		c.String(http.StatusBadRequest, "Ugyldig epost")
		return
	} else if v[1] != "soprasteria.com" {
		log.Printf("User with email-domain: %s tried to sign up.", v[1])
		c.String(http.StatusBadRequest, "Beklager, du kan ikke registrere deg (enda)")
		return
	}

	log.Printf("Username: %s \n", user.Username)
	log.Printf("Email: %s \n", user.Email)

	//Is username used before?
	usernameExists, err := timerDb.UserExistsWithUsername(user.Username)
	if usernameExists {
		log.Printf("User with username: %s Already exists.", user.Username)
		//TODO: Give better response here.
		c.String(http.StatusUnprocessableEntity, "Brukernavn %s er allerede tatt", user.Username)
		return
	}
	if err != nil {
		log.Printf("Noe gikk galt under DB kall %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	//Is email used before?
	emailExists, _, err := timerDb.UserExistsWithEmail(user.Email)
	if emailExists {
		log.Printf("Email: %s is already in use.", user.Email)
		//TODO: Give better response here? Shouldn't inform that this email is in use. Makes scraping possible
		c.String(http.StatusBadRequest, "Noe gikk galt. Prøv igjen senere")
		return
	}
	if err != nil {
		log.Printf("Noe gikk galt under DB kall:\n%s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	//Create user
	userid, err := timerDb.CreateUser(user)
	if err != nil {
		c.String(http.StatusInternalServerError, "Noe gikk galt under lagring av brukereren. Prøv på nytt senere")
		log.Printf("Error on create: \n%s", err.Error())
		return
	}

	err = sendAuthMail(userid, user.Email)
	if err != nil {
		log.Printf("Something went wrong sending email to user. \n%s", err.Error())
		c.String(http.StatusInternalServerError, "Noe gikk galt under utsendelse av bekreftelses e-post. Forsøk å logge inn med din epost adresse på nytt", nil)
		return
	}

	c.Header("Location", "/autentisering/engangskode")
	c.Status(http.StatusSeeOther)
}

func sendAuthMail(userId int64, email string) error {
	res, err := timerDb.GetUser(userId)
	if err != nil {
		log.Print("Could not retrieve")
		log.Print(err.Error())
		return err
	}

	templateId := "d-67cb50f335c44f85a8960612dc97e7bc"
	httpposturl := "https://api.sendgrid.com/v3/mail/send"

	postString := fmt.Sprintf(`{
		"from":{
			"email":"kim.nilsenbrusevold@outlook.com"
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
	}`, email, res.OneTimeCode.String, templateId)

	postdata := []byte(postString)
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(postdata))
	if err != nil {
		log.Printf("Klarte ikke å forberede request for å sende template. %s", err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json;")
	sendgridApiKey, exists := os.LookupEnv("SENDGRID_API_KEY")
	if !exists {
		log.Print("No env variable named 'SENDGRID_API_KEY' in .env file or environment variable. Exiting")
		return errors.New("no env variable named 'SENDGRID_API_KEY' found. Cannot send email")
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sendgridApiKey))

	log.Printf("Forsøker å sende epost til bruker med epost %s", email)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error when sending email request to SendGrid. \n %s", err.Error())
		return err
	}
	defer response.Body.Close()

	log.Printf("Sendgripd response Status: %s", response.Status)
	responseCont, err := io.ReadAll(response.Body)
	if err != nil {
		log.Print("Cannot read response body")
	} else {
		log.Print(string(responseCont))
	}
	return nil
}

func validateEmail(c *gin.Context) {
	log.Printf("Validating email")
	email := c.PostForm("email")
	emailParts := strings.Split(strings.TrimSpace(email), "@")

	if len(emailParts) != 2 || emailParts[1] != "soprasteria.com" {
		returnForm := `
		<div hx-target="this" hx-swap="outerHTML">
        	<label for="email">Din epost: </label>
        	<input hx-post="/registrer-bruker/email" type="email" name="email" id="email" value="%s" required />
			<div class='error-message'>Ugyldig epost. Kun Sopra Steria kan registrere seg på dette tidspunktet</div>
      	</div>
		`
		c.String(http.StatusOK, returnForm, email)
		log.Printf("Invalid email %s", email)
		log.Printf("Length: %d", len(emailParts))
		return
	}
	returnForm := `
		<div hx-target="this" hx-swap="outerHTML">
        	<label for="email">Din epost: </label>
        	<input hx-post="/registrer-bruker/email" type="email" name="email" id="email" value="%s" required />
      	</div>
		`
	c.String(http.StatusOK, returnForm, email)
}

func validateUsername(c *gin.Context) {
	username := c.PostForm("username")
	username = strings.TrimSpace(username)
	usernameInUse, err := timerDb.UserExistsWithUsername(username)
	if err != nil {
		log.Printf("Error getting user with username %s. Error: %s", username, err.Error())
		c.String(http.StatusInternalServerError, "")
		return
	}

	var returnform string
	if usernameInUse {
		returnform = `
		<div hx-target="this" hx-swap="outerHTML">
			<label for="username">Ønsket Brukernavn</label>
			<input hx-post="/registrer-bruker/username" type="text" name="username" id="username" value="%s" required />
			<div class='error-message'>Dette brukernavnet er ikke tilgjengelig</div>
		</div>
		`
	} else {
		returnform = `
			<div hx-target="this" hx-swap="outerHTML">
				<label for="username">Ønsket Brukernavn</label>
				<input hx-post="/registrer-bruker/username" type="text" name="username" id="username" value="%s" required />
			</div>
			`
	}
	c.String(http.StatusOK, returnform, username)
}
