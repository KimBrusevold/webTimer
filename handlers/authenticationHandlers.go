package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleAuthentication(rg *gin.RouterGroup) {
	rg.GET("/login", loginPage)
	rg.POST("/login", loginUser)

}

func loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func loginUser(c *gin.Context) {
	email := c.PostForm("email")

	if email == "" {
		log.Print("Could not bind form data to user")
		c.String(http.StatusBadRequest, "Ugyldig epost eller navn")
		return
	}

	usernameExists, id, err := timerDb.UserExistsWithEmail(email)

	if err != nil || !usernameExists {
		log.Printf("Kunne ikke logge inn bruker. Bruker finnes: %t. DB feil: %s", usernameExists, err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	_, err = timerDb.SetNewOnetimeCode(email)

	if err != nil {
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	err = sendAuthMail(id, email)
	if err != nil {
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
}
