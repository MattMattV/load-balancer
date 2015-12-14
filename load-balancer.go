package main

import (
	"errors"
	"github.com/google/cadvisor/client"
	info "github.com/google/cadvisor/info/v1"
	"io"
	"log"
	"net"
	"runtime"
	"strings"
)

var filter string = "dummy"
var monitor string = "http://cadvisor:8080"
var port string = ":6666"

// determine the container with the more available RAM
func getLessLoaded() (string, error) {

	allContainers, err := getAllContainerInfo(monitor)

	if detectError(err, false) {
		return "", err
	}

	var alias, ret string
	var kbFree uint64

	// will contain for each container name, its available RAM
	mapContainers := make(map[string]uint64)

	for _, item := range allContainers {
		alias = item.Aliases[0]
		kbFree = item.Stats[0].Memory.Usage

		mapContainers[alias] = kbFree
	}

	var isItemFound bool = false

	for key, _ := range mapContainers {

		// finding the first container to look for the next comparison
		if isItemFound == false {
			if strings.Contains(key, filter) {
				ret = key
				isItemFound = true
			}
		}

		// verifying only wanted containers
		if strings.Contains(key, filter) {
			if mapContainers[key] < mapContainers[ret] {
				ret = key
			}
		}
	}

	if ret == "" {
		return "", errors.New("No server found...")
	}
	return ret, nil
}

func getAllContainerInfo(cadvisor string) ([]info.ContainerInfo, error) {

	client, err := client.NewClient(cadvisor)
	if detectError(err, true) {
		return nil, err
	}

	request := info.DefaultContainerInfoRequest()
	allContainers, err := client.AllDockerContainers(&request)
	if detectError(err, true) {
		return nil, err
	}

	return allContainers, nil
}

func detectError(err error, doLog bool) bool {

	if err != nil {
		if doLog {
			log.Println("LOG: ", err)
		}
		return true
	} else {
		return false
	}
}

// will attempt to encapsulate all the TCP traffic by linking the sockets together
func forwarding(client net.Conn) {

	lessLoaded, err := getLessLoaded()
	detectError(err, true)

	target, err := net.Dial("tcp", lessLoaded+":6379")
	detectError(err, true)

	//copying the client Writer to the target Reader
	go func() {
		_, err := io.Copy(target, client)
		detectError(err, true)
	}()

	//copying the client Reader to the target Writer
	go func() {
		_, err := io.Copy(client, target)
		detectError(err, true)
	}()
}

func main() {

	log.Println("Listening on http://127.0.0.1" + port)
	ln, err := net.Listen("tcp", port)
	detectError(err, true)

	for {
		conn, err := ln.Accept()
		log.Println("Accept? ", err)
		go forwarding(conn)
		log.Println("nb goroutines : ", runtime.NumGoroutine())
	}
}
