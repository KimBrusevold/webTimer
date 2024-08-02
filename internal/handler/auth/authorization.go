package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ah AuthHandler) loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"title": "Logg inn",
	})
}

func (ah AuthHandler) loginUser(c *gin.Context) {
	email := c.PostForm("email")

	if email == "" {
		log.Print("Could not bind form data to user")
		c.String(http.StatusBadRequest, "Ugyldig epost eller navn")
		return
	}

	usernameExists, _, err := ah.DB.UserExistsWithEmail(email)

	if err != nil {
		log.Printf("Kunne ikke logge inn bruker. DB feil: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !usernameExists {
		log.Printf("Exists no user with email address: %s", email)
		c.Header("Location", "/registrer-bruker")
		c.Status(http.StatusSeeOther)
		return
	}
	password := c.PostForm("password")
	user, err := ah.DB.UserAuthProcess(email, password)
	if err != nil {
		c.String(http.StatusUnauthorized, "Error on authentication: %s", err.Error())
		return
	}

	c.SetCookie("userAuthCookie", user.Authcode.String, 0, "/", hostUrl, true, true)
	c.SetCookie("userId", fmt.Sprintf("%d", user.ID), 0, "/", hostUrl, true, true)

	c.Header("Location", "/")
	c.Status(http.StatusSeeOther)
	// _, err = timerDb.SetNewOnetimeCode(email)

	// if err != nil {
	// 	log.Print(err.Error())
	// 	c.Status(http.StatusInternalServerError)
	// 	return
	// }

	// err = sendAuthMail(id, email)
	// if err != nil {
	// 	log.Print(err.Error())
	// 	c.Status(http.StatusInternalServerError)
	// 	return
	// }
	// c.Header("Location", "/autentisering/engangskode")
	// c.Status(http.StatusSeeOther)
}

// func onetimeCode(c *gin.Context) {
// 	c.HTML(http.StatusOK, "email-sent.html", nil)
// }
