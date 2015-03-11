package main

type Queue struct {
	ID     int
	Tracks []Track
}

var idToList map[int]*Queue

func init() {
	idToList = make(map[int]*Queue)
}

func NewQueue(id int) *Queue {
	q := new(Queue)
	q.ID = id
	q.Tracks = make([]Track, 0)
	return q
}
