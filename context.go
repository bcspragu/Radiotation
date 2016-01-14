package main

import (
	"net/http"
	"room"

	"golang.org/x/net/context"
)

type Context struct {
	context.Context

	w     http.ResponseWriter
	r     *http.Request
	User  *room.User
	Room  *room.Room
	Queue *room.Queue
}

func NewContext(w http.ResponseWriter, r *http.Request) Context {
	return Context{
		w:       w,
		r:       r,
		Context: context.Background(),
	}
}

func (c Context) adminError(err error) {
	c.w.Write([]byte("You had an error brah: " + err.Error()))
}
