package main

import (
	"github.com/gorilla/securecookie"
	"io/ioutil"
)

func main() {
	ioutil.WriteFile("hashKey", securecookie.GenerateRandomKey(32), 0644)
	ioutil.WriteFile("blockKey", securecookie.GenerateRandomKey(32), 0644)
}
