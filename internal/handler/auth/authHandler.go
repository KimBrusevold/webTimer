package auth

import (
	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	DB *database.TimerDB
}

func (a AuthHandler) SetupRoutes(rg *gin.RouterGroup) {
	rg.GET("/innlogging", a.loginPage)
	rg.POST("/innlogging", a.loginUser)

	rg.GET("/registrer-bruker", a.registerUserPage)
	rg.POST("/registrer-bruker", a.createUser)
}
