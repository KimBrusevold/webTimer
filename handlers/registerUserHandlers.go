package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/KimBrusevold/webTimer/timer"
	"github.com/gin-gonic/gin"
)

func HandleRegisterUser(rg *gin.RouterGroup) {
	rg.GET("/", registerUserPage)
	rg.POST("/", createUser)
}

func registerUserPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register-user.html", nil)
}

func createUser(c *gin.Context) {
	user := timer.User{
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
	}

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
	}
	if v[1] != "soprasteria.com" {
		log.Printf("User with email-domain: %s tried to sign up.", v[1])
		c.String(http.StatusBadRequest, "Beklager, du kan ikke registrere deg (enda)")
		return
	}

	log.Printf("Username: %s \n", user.Username)
	log.Printf("Email: %s \n", user.Email)

	c.String(http.StatusOK, "Username: %s, Email: %s", user.Username, user.Email)
	return

	// user := timer.User{
	// 	Username: username,
	// 	Email:    email,
	// }

	// userid, err := timerDb.CreateUser(user)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	log.Print("Error on create: \n")
	// 	log.Print(err.Error())
	// 	return
	// }

	// err = SendAuthMail(userid, email)
	// if err != nil {
	// 	w.Write([]byte("Something went wrong"))
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// content, err := os.ReadFile("./pages/email-sent.html")
	// if err != nil {
	// 	log.Print("ERROR: Could not read email-sent.html from file")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// _, err = w.Write(content)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

}
