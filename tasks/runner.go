package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

var kill = make(chan struct{})
var gin *exec.Cmd
var stdout io.Reader
var err error

func main() {
	gin = exec.Command("gin", "--appPort", "8000", "--port", "8080", "--filetype=go,html", "--exclude=.git,bower_components,node_modules", "--", "--env=development")

	ginOut, err := gin.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := gin.Start(); err != nil {
		log.Fatal(err)
	}

	go restart()

	r := bufio.NewScanner(ginOut)
	for r.Scan() {
		if strings.Contains(r.Text(), "proxy") {
			kill <- struct{}{}
		} else {
			fmt.Println(r.Text())
		}
	}
}

func restart() {
	for {
		<-kill
		fmt.Println("Killing and restarting")
		gin.Process.Kill()
		gin = exec.Command("gin", "--appPort", "8000", "--port", "8080", "--filetype=go,html", "--exclude=.git,bower_components,node_modules")
		stdout, err = gin.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := gin.Start(); err != nil {
			log.Fatal(err)
		}
	}
}
