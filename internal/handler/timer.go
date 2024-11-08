package handler

import (
	"log"
	"net/http"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/KimBrusevold/webTimer/internal/middelware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TimerHandler struct {
	DB *database.TimerDB
}

func (th TimerHandler) SetupRoutes(rg *gin.RouterGroup) {
	authMW := middelware.AuthMiddelware{
		DB: th.DB,
	}
	rg.Use(authMW.Authenticate)
	rg.GET("/start-lop", getTimerPage)
	rg.POST("/start-lop", th.startTimerHandler)
	rg.GET("/avslutt-lop", th.endTimerHandler)
}

func getTimerPage(c *gin.Context) {
	uid := uuid.New().String()
	c.HTML(http.StatusOK, "start-tid.tmpl", gin.H{
		"uid": uid,
	})
}

func (th TimerHandler) startTimerHandler(c *gin.Context) {
	i, exists := c.Get("userId")

	if !exists {
		log.Print("Found no userId in context. Cannot start timer")
		c.Status(http.StatusInternalServerError)
		return
	}

	randomString, err := th.DB.StartTimer(i.(int))

	if err != nil {
		log.Print("Could not start timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.SetCookie("otc-start", randomString, 6000, "/timer", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{
		"otc": randomString,
	})
}

func (th TimerHandler) endTimerHandler(c *gin.Context) {
	i, exists := c.Get("userId")
	if !exists {
		log.Print("Found no userId in context. Cannot start timer")
		c.Status(http.StatusInternalServerError)
		return
	}

	timeUsed, err := th.DB.EndTimeTimer(i.(int))
	if err != nil {
		log.Print("Could not stop timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	minutes := timeUsed / (60 * 1000) % 60
	seconds := timeUsed / (1000) % 60
	tenths := timeUsed / (100) % 1000
	c.HTML(http.StatusOK, "tid-avsluttet.tmpl", gin.H{
		"minutes": minutes,
		"seconds": seconds,
		"tenths":  tenths,
	})

}
