package main

import (
	info "github.com/google/cadvisor/info/v1"

	"errors"
	"fmt"
	"github.com/google/cadvisor/client"
	"log"
	"net/http"
	"strings"
)

var filter string = "dummy"
var monitor string = "http://cadvisor:8080"
var port string = ":9999"

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

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("GET vue")
	server, err := getLessLoaded()
	log.Println("got server : ", server)

	if detectError(err, false) {
		w.WriteHeader(500) //warn the user that the server encountered a problem
	} else {
		w.WriteHeader(200)
	}

	w.Write([]byte(server))
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
			log.Println(err)
		}
		return true
	} else {
		return false
	}
}

func main() {

	http.HandleFunc("/", handleRoot)

	fmt.Println("Listening on http://127.0.0.1" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
