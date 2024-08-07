package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/gin-gonic/gin"
)

var hostUrl string

type fieldResponse struct {
	Error        bool
	Errormessage string
	Value        string
}

func (ah AuthHandler) registerUserPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register-user.tmpl", gin.H{
		"email": fieldResponse{
			Error: false,
		},
	})
}

func (ah AuthHandler) createUser(c *gin.Context) {
	user := database.User{
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
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
	usernameExists, err := ah.DB.UserExistsWithUsername(user.Username)
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
	emailExists, _, err := ah.DB.UserExistsWithEmail(user.Email)
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
	_, err = ah.DB.CreateUser(user)
	if err != nil {
		c.String(http.StatusInternalServerError, "Noe gikk galt under lagring av brukereren. Prøv på nytt senere")
		log.Printf("Error on create: \n%s", err.Error())
		return
	}

	err = ah.EmailClient.SendAuthEmail(user.Email)
	if err != nil {
		log.Printf("Error sending email: %s", err)
	}
	c.Header("Location", "/aut/innlogging")
	c.Status(http.StatusSeeOther)
}

// func validateEmail(c *gin.Context) {
// 	log.Printf("Validating email")
// 	email := c.PostForm("email")
// 	emailParts := strings.Split(strings.TrimSpace(email), "@")

// 	if len(emailParts) != 2 || emailParts[1] != "soprasteria.com" {
// 		c.HTML(http.StatusOK, "emailfield", fieldResponse{
// 			Error:        true,
// 			Errormessage: "Ugyldig epost. Kun Sopra Steria kan registrere seg på dette tidspunktet",
// 			Value:        email,
// 		})
// 		log.Printf("Invalid email %s", email)
// 		log.Printf("Length: %d", len(emailParts))
// 		return
// 		// return

// 	}
// 	c.HTML(http.StatusOK, "emailfield", fieldResponse{
// 		Error: false,
// 		Value: email,
// 	})
// }

// func validatePassword(c *gin.Context) {
// 	p := c.PostForm("password")
// 	if len(p) < 5 {
// 		returnForm := `
// 		<div hx-target="this" hx-swap="outerHTML">
//         	<label for="passord">Ditt passord: </label>
//         	<input hx-post="/registrer-bruker/password" type="password" name="password" id="password" value="%s" required />
// 			<div class='error-message'>Passorder er for kort</div>
//       	</div>
// 		`
// 		c.String(http.StatusOK, returnForm, p)
// 		return
// 	} else if len([]byte(p)) > 72 {
// 		returnForm := `
// 		<div hx-target="this" hx-swap="outerHTML">
//         	<label for="passord">Ditt passord: </label>
//         	<input hx-post="/registrer-bruker/password" type="password" name="password" id="password" value="%s" required />
// 			<div class='error-message'>Passordet er for kort.</div>
//       	</div>
// 		`
// 		c.String(http.StatusOK, returnForm, p)
// 	}
// 	returnForm := `
// 	<div hx-target="this" hx-swap="outerHTML">
// 		<label for="passord">Ditt passord: </label>
// 		<input hx-post="/registrer-bruker/password" type="password" name="password" id="password" value="%s" required />
// 	</div>
// 		`
// 	c.String(http.StatusOK, returnForm, p)
// }

// func (ah AuthHandler) validateUsername(c *gin.Context) {
// 	username := c.PostForm("username")
// 	username = strings.TrimSpace(username)
// 	usernameInUse, err := ah.DB.UserExistsWithUsername(username)
// 	if err != nil {
// 		log.Printf("Error getting user with username %s. Error: %s", username, err.Error())
// 		c.String(http.StatusInternalServerError, "")
// 		return
// 	}

// 	var returnform string
// 	if usernameInUse {
// 		returnform = `
// 		<div hx-target="this" hx-swap="outerHTML">
// 			<label for="username">Ønsket Brukernavn</label>
// 			<input hx-post="/registrer-bruker/username" type="text" name="username" id="username" value="%s" required />
// 			<div class='error-message'>Dette brukernavnet er ikke tilgjengelig</div>
// 		</div>
// 		`
// 	} else {
// 		returnform = `
// 			<div hx-target="this" hx-swap="outerHTML">
// 				<label for="username">Ønsket Brukernavn</label>
// 				<input hx-post="/registrer-bruker/username" type="text" name="username" id="username" value="%s" required />
// 			</div>
// 			`
// 	}
// 	c.String(http.StatusOK, returnform, username)
// }
