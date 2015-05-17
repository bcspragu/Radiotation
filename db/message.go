package db

import (
	"fmt"
)

func RecentMessages() []string {
	var msgs []string
	rows, err := d.Query("SELECT body FROM (SELECT body, created FROM messages ORDER BY created DESC LIMIT 100) AS recent ORDER BY created")
	if err != nil {
		fmt.Println(err)
		return msgs
	}

	defer rows.Close()
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			fmt.Println(err)
			return msgs
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return msgs
	}
	return msgs
}
