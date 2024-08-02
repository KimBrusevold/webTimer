package handler

import (
	"log"
	"net/http"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/KimBrusevold/webTimer/internal/model"
	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	DB *database.TimerDB
}

func (lh LeaderboardHandler) HandleLeaderboardShow(c *gin.Context) {
	times, err := lh.DB.RetrieveTimes()
	if err != nil {
		log.Printf("Could not get times from db. %s", err.Error())
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	var timesDisplay []model.TimesDisplay

	for _, t := range times {
		td := model.TimesDisplay{
			Place:    t.Place,
			Username: t.Username,
			Minutes:  t.ComputedTime / (60 * 1000) % 60,
			Seconds:  t.ComputedTime / (1000) % 60,
			Tenths:   t.ComputedTime / (100) % 1000,
		}
		timesDisplay = append(timesDisplay, td)
	}

	c.HTML(http.StatusOK, "leaderboard.tmpl", gin.H{
		"title": "Resultatliste",
		"data":  timesDisplay,
	})
}
