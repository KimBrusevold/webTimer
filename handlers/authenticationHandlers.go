package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleAuthentication(rg *gin.RouterGroup) {
	rg.GET("/login", loginPage)
	rg.POST("/login", loginUser)

	rg.GET("/emailredirect/:authcode", authenticateEmailCode)
}

func authenticateEmailCode(c *gin.Context) {
	authcode := c.Param("authcode")

	user, err := timerDb.UserAuthProcess(authcode)
	if err != nil {
		c.String(http.StatusUnauthorized, "Error on authentication: %s", err.Error())
		return
	}
	c.SetCookie("userAuthCookie", user.Authcode.String, 0, "/", hostUrl, true, true)
	c.SetCookie("userId", fmt.Sprintf("%d", user.ID), 0, "/", hostUrl, true, true)

	c.String(http.StatusOK, "Authenticated at %s. Hello %s", authcode, user.Username)

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
