package main

import (
	info "github.com/google/cadvisor/info/v1"

	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/cadvisor/client"
	"os"
	"sync"
	"time"
)

// monitor : URL where the cAdvisor container is
// filter  : the name of the servers the clients want to contact
// port    : port we will listen on
var filter, monitor, port string
var mapContainers = make(map[string]uint64)
var mutex sync.Mutex

func updateContainers() {

	listContainers()

	for _ = range time.Tick(3 * time.Second) {
		log.Println("Starting list of containers update...")
		listContainers()
		log.Print("Done.\n\n")
	}
}

func listContainers() {

	allContainers, _ := getAllContainerInfo(monitor)

	// resetting mapContainers
	mutex.Lock()
	for key := range mapContainers {
		delete(mapContainers, key)
	}

	// filtering data and filling mapContainers
	for _, item := range allContainers {
		alias := item.Aliases[0]
		kbFree := item.Stats[0].Memory.Usage

		if strings.Contains(alias, filter) {
			mapContainers[alias] = kbFree
		}
	}

	log.Println("\tFound", len(mapContainers), "containers")
	for key, value := range mapContainers {
		log.Printf("\t\t%s %d", key, value)
	}
	mutex.Unlock()
}

// determine the container with the more available RAM
func getLessLoaded() (string, error) {

	var lessLoaded string

	mutex.Lock()
	for key := range mapContainers {

		if lessLoaded == "" {
			lessLoaded = key
		} else if mapContainers[key] < mapContainers[lessLoaded] {
			lessLoaded = key
		}
	}
	mutex.Unlock()

	if lessLoaded == "" {
		return "", errors.New("No server found...")
	}
	return lessLoaded, nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	server, err := getLessLoaded()

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
	}
	return false
}

func main() {

	filter = os.Getenv("FILTER")
	monitor = os.Getenv("MONITOR")
	port = os.Getenv("HTTP_PORT")

	if filter == "" {
		log.Fatalln("FILTER environment variable is missing")
	}
	if monitor == "" {
		log.Fatalln("MONITOR environment variable is missing")
	}

	if port == "" {
		log.Fatalln("HTTP_PORT environment variable is missing")
	}

	go updateContainers()

	http.HandleFunc("/", handleRoot)

	fmt.Println("Listening on http://127.0.0.1" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
