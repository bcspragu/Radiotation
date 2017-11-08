package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/namsral/flag"
)

var (
	kill = make(chan struct{})

	port, appPort int
	gin           *exec.Cmd
	stdout        io.Reader
	err           error
)

func main() {
	flag.IntVar(&appPort, "app_port", 8000, "The port to run the backend on")
	flag.IntVar(&port, "port", 8080, "The port to run gin on and to use in the browser")
	flag.Parse()

	gin = exec.Command("gin", "--appPort", strconv.Itoa(appPort), "--port", strconv.Itoa(port), "--filetype=go", "--excludeDir=.git,bower_components,node_modules,frontend")

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
		gin = exec.Command("gin", "--appPort", strconv.Itoa(appPort), "--port", strconv.Itoa(port), "--filetype=go,html", "--excludeDir=.git,bower_components,node_modules")
		stdout, err = gin.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := gin.Start(); err != nil {
			log.Fatal(err)
		}
	}
}
