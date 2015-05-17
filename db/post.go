package db

import (
	"database/sql"
	"fmt"
)

type Post struct {
	ID       int
	Body     string
	Comments []string
}

func RecentPosts() []*Post {
	postCheck := make(map[int]*Post)
	var posts []*Post
	rows, err := d.Query("SELECT posts.id, posts.body, comments.body FROM posts FULL OUTER JOIN comments ON posts.id = comments.post_id ORDER BY posts.created, comments.created")
	if err != nil {
		fmt.Println("Error executing query:", err)
		return []*Post{}
	}

	defer rows.Close()
	for rows.Next() {
		var p = new(Post)
		var comment sql.NullString
		if err := rows.Scan(&p.ID, &p.Body, &comment); err != nil {
			fmt.Println("Error putting data into fields", err)
			return []*Post{}
		}
		if post, ok := postCheck[p.ID]; ok {
			if comment.Valid {
				post.Comments = append(post.Comments, comment.String)
			}
		} else {
			postCheck[p.ID] = p
			posts = append(posts, p)
			if comment.Valid {
				p.Comments = []string{comment.String}
			} else {
				p.Comments = []string{}
			}
		}
	}
	return posts
}
