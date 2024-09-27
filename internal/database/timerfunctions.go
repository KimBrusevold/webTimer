package database

import (
	"log"
	"time"
)

type TimesCountRespose struct {
	Place    int
	Count    int
	Username string
}

func (r *TimerDB) RetrieveTimesCount() ([]TimesCountRespose, error) {
	query := `SELECT ROW_NUMBER () OVER (ORDER BY Count(t.id) DESC), Count(t.id), users.username FROM times t
 INNER JOIN  users on users.id = t.userid GROUP BY userid;`
	rows, err := r.db.Query(query)
	log.Print("Queried database")
	if err != nil {
		log.Printf("database query failed %s", err)

		return nil, err
	}
	defer rows.Close()
	var times []TimesCountRespose

	for rows.Next() {
		var tim TimesCountRespose
		if err := rows.Scan(&tim.Place, &tim.Count, &tim.Username); err != nil {
			return times, err
		}

		times = append(times, tim)
	}

	if err = rows.Err(); err != nil {
		return times, err
	}

	return times, nil
}

// Get fastest times by times. Time provided should be an UTC date.
func (r *TimerDB) RetrieveFastestTimeByTime(from time.Time,to time.Time) ([]RetrieveTimesResponse, error) {

	query := `SELECT ROW_NUMBER () OVER (ORDER BY times.computedtime ASC) rownum, min(times.computedtime), username FROM times 
		INNER JOIN users on users.id = userid
		WHERE times.computedtime IS NOT NULL
		AND times.starttime >= ?
		AND times.startTime < ?
		GROUP BY userid;`
	rows, err := r.db.Query(query, from.UnixMilli(), to.UnixMilli())
	if err != nil {	
		log.Printf("database query failed %s", err)
		return nil, err
	}
	defer rows.Close()

	var times []RetrieveTimesResponse

	for rows.Next() {
		var tim RetrieveTimesResponse
		if err := rows.Scan(&tim.Place, &tim.ComputedTime, &tim.Username); err != nil {
			return times, err
		}

		times = append(times, tim)
	}

	if err = rows.Err(); err != nil {
		return times, err
	}

	return times, nil
}