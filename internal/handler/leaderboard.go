package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/KimBrusevold/webTimer/internal/database"
	"github.com/KimBrusevold/webTimer/internal/model"
	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	DB *database.TimerDB
}

func (lh LeaderboardHandler) HandleLeaderboardShow(c *gin.Context) {
	times, err := lh.DB.RetrieveFastestTimeByTime(getRangeToday())
	if err != nil {
		log.Printf("Could not get times from db. %s", err.Error())
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	number, err := lh.DB.RetrieveTimesCount()
	if err != nil {
		log.Printf("Could not get times count from db. %s", err.Error())
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
		"title":      "Resultatliste",
		"timingData": timesDisplay,
		"countData":  number,
	})
}

func (lh LeaderboardHandler)RenderFastestLeaderboard(c *gin.Context) {
	filter := c.DefaultQuery("filter", "idag")
	log.Printf("QUERY AT %s", filter)
	var times []database.RetrieveTimesResponse
	var err error

	if(filter == "idag") {
		times, err = lh.DB.RetrieveFastestTimeByTime(getRangeToday())
	} else if filter == "noensinne" {
		times, err = lh.DB.RetrieveAllTimeFastestTimes()
	}
	if err != nil {
		log.Printf("Error getting fastest time %s", err)
	}
	log.Printf("FANT: %d tider", len(times))

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

	c.HTML(http.StatusOK, "leaderboardTable.tmpl", gin.H {
		"leaderboardOfHeader": "Tid",
		"timingData": timesDisplay,
	})
}

func getRangeToday() (time.Time, time.Time) {
	now := time.Now().UTC()
	currYear, currMont, currDay := now.Date()
	startOfDay := time.Date(currYear, currMont, currDay,0,0,0,0, now.Location())
	startOfTomorow := startOfDay.AddDate(0,0,1)

	log.Printf("fra: %d \n til: %d", startOfDay.UnixMilli(), startOfTomorow.UnixMilli())
	return startOfDay, startOfTomorow
}

func getRangeCurrentMonth() (time.Time, time.Time) {
	now := time.Now().UTC()
	currYear, currMont, _ := now.Date()
	firstOfMonth := time.Date(currYear, currMont, 1,0,0,0,0, now.Location())
	nextMonth := firstOfMonth.AddDate(0,1,0)

	return  firstOfMonth, nextMonth
}
