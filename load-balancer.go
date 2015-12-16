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

// for every tick, will ask to update the data with the running containers
// this method will run in background so it will not stop the HTTP server
func updateContainers() {

	listContainers()

	for _ = range time.Tick(3 * time.Second) {
		log.Println("Starting list of containers update...")
		listContainers()
		log.Print("Done.\n\n")
	}
}

// update the mapContainers variable with containers containing the filter variable in their name
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

	// logging for the containers in mapContainers
	log.Println("\tFound", len(mapContainers), "containers")
	for key, value := range mapContainers {
		log.Printf("\t\t%s %d", key, value)
	}
	mutex.Unlock()
}

// determine the container with the more available RAM
// throw an error when no servers are found
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

// will write to the client the less loaded server
// the reply will include an HTTP code between 200 (ok) and 500 (server encountered an error)
func handleRoot(w http.ResponseWriter, r *http.Request) {
	server, err := getLessLoaded()

	if detectError(err, false) {
		w.WriteHeader(500) //warn the user that the server encountered a problem
	} else {
		w.WriteHeader(200)
	}

	w.Write([]byte(server))
}

// query cAdvisor for all the running containers on the same host
// throw an error if cAdvisor get one, it will only log it in standard output
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

// will help to display any error
func detectError(err error, doLog bool) bool {

	if err != nil {
		if doLog {
			log.Println(err)
		}
		return true
	}
	return false
}

// clients will be able to ask at / which server is the less loaded
func main() {

	// getting all the variables needed to run
	filter = os.Getenv("FILTER")
	monitor = os.Getenv("MONITOR")
	port = os.Getenv("HTTP_PORT")

	// check if all variables are set
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
