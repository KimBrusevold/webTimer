package handler

import (
	"log"
	"net/http"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/KimBrusevold/webTimer/internal/middelware"
	"github.com/gin-gonic/gin"
)

type TimerHandler struct {
	DB *database.TimerDB
}

func (th TimerHandler) SetupRoutes(rg *gin.RouterGroup) {
	authMW := middelware.AuthMiddelware{
		DB: th.DB,
	}
	rg.Use(authMW.Authenticate)
	rg.GET("/start-lop", th.startTimerHandler)
	rg.GET("/avslutt-lop", th.endTimerHandler)
}

func (th TimerHandler) startTimerHandler(c *gin.Context) {
	i, exists := c.Get("userId")
	if !exists {
		log.Print("Found no userId in context. Cannot start timer")
		c.Status(http.StatusInternalServerError)
		return
	}
	err := th.DB.StartTimer(i.(int))
	if err != nil {
		log.Print("Could not start timer")
		log.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "tid-startet.tmpl", nil)
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
