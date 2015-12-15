package main

import (
	info "github.com/google/cadvisor/info/v1"

	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/cadvisor/client"
	"time"
)

var filter string = "dummy"                  //the name of the servers the clients want to contact
var monitor string = "http://127.0.0.1:8080" // URL where the cAdvisor container is
var port string = ":6666"                    // port we will listen on
var mapContainers = make(map[string]uint64)

func updateContainers() {

	listContainers()

	for _ = range time.Tick(5 * time.Second) {
		log.Println("Starting list of containers update...")
		listContainers()
		log.Println("Found", len(mapContainers), "containers")
		log.Println("Done.\n")
	}
}

func listContainers() {

	allContainers, _ := getAllContainerInfo(monitor)

	var alias string
	var kbFree uint64

	// resetting mapContainers
	for key, _ := range mapContainers {
		delete(mapContainers, key)
	}

	for _, item := range allContainers {
		alias = item.Aliases[0]
		kbFree = item.Stats[0].Memory.Usage

		if strings.Contains(alias, filter) {
			mapContainers[alias] = kbFree
		}
	}
}

// determine the container with the more available RAM
func getLessLoaded() (string, error) {

	var lessLoaded string

	for key, _ := range mapContainers {
		lessLoaded = key
		break
	}

	for key, _ := range mapContainers {

		// verifying only wanted containers
		if mapContainers[key] < mapContainers[lessLoaded] {
			lessLoaded = key
		}
	}

	if lessLoaded == "" {
		return "", errors.New("No server found...")
	}
	return lessLoaded, nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("GET vue")
	server, err := getLessLoaded()
	log.Println("chose server : ", server)

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

	go updateContainers()

	http.HandleFunc("/", handleRoot)

	fmt.Println("Listening on http://127.0.0.1" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
