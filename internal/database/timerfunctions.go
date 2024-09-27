package database

type TimesCountRespose struct {
	Place    int
	Count    int
	Username string
}

func (r *TimerDB) RetrieveTimesCount() ([]TimesCountRespose, error) {
	query := `SELECT ROW_NUMBER () OVER (ORDER BY Count(t.id) DESC), Count(t.id), users.username FROM times t
 INNER JOIN  users on users.id = t.userid GROUP BY userid;`
	rows, err := r.db.Query(query)
	if err != nil {
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
