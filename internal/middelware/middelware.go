package middelware

import (
	"log"
	"net/http"
	"strconv"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/gin-gonic/gin"
)

type AuthMiddelware struct {
	DB *database.TimerDB
}

func (amw *AuthMiddelware) Authenticate(c *gin.Context) {
	userCookie, cErr := c.Cookie("userAuthCookie")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userAuthCookie")
		log.Print(cErr.Error())
		c.Header("Location", "/aut/innlogging")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	idCookie, cErr := c.Cookie("userId")
	if cErr != nil {
		log.Print("User not authenticated. Does not have userId cookie")
		log.Print(cErr.Error())
		c.Header("Location", "/aut/innlogging")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	i, err := strconv.Atoi(idCookie)
	if err != nil {
		log.Print("Could not get id from cookie")
		log.Print(err.Error())
		c.Header("Location", "/aut/innlogging")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	isAuthenticated := amw.DB.IsAuthorizedUser(userCookie, i)
	if !isAuthenticated {
		log.Print("Could not find user with id and auth code")
		c.Header("Location", "/aut/innlogging")
		c.Status(http.StatusSeeOther)
		c.Abort()
		return
	}

	c.Set("userId", i)
}
