package auth

import (
	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/KimBrusevold/webTimer/internal/email"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	DB          *database.TimerDB
	EmailClient *email.EmailClient
}

func (a AuthHandler) SetupRoutes(rg *gin.RouterGroup) {
	rg.GET("/innlogging", a.loginPage)
	rg.POST("/innlogging", a.loginUser)

	rg.GET("/registrer-bruker", a.registerUserPage)
	rg.POST("/registrer-bruker", a.createUser)

	rg.POST("/engangskode", a.oneTimeCode)

	rg.GET("/nytt-passord", a.newPassword)
	rg.POST("/nytt-passord", a.setnewPassword)

	rg.POST("/nytt-passord/email", a.sendNewPasswordEmail)
}
