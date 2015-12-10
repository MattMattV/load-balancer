package main

import (
	"github.com/google/cadvisor/client"

	info "github.com/google/cadvisor/info/v1"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var filter string = "dummy"
var monitor string = "http://cadvisor:8080"
var middleware string = "http://middleware:9090/request/"
var port string = ":7000"

// determine the container with the more available RAM
func getLessLoaded() string {

	allContainers := getAllContainerInfo(monitor)

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

	return ret
}

func handleView(w http.ResponseWriter, r *http.Request) {
	log.Println("GET vue")
	w.Write([]byte(getLessLoaded()))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("GET racine")
	w.Write([]byte("Hello World !"))
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	log.Println("GET redirige")

	lessLoaded := getLessLoaded()
	resp := askMiddleware(lessLoaded)
	w.Write(resp)
}

func getAllContainerInfo(cadvisor string) []info.ContainerInfo {

	client, err := client.NewClient(cadvisor)
	if detectError(err) {
		return nil
	}

	request := info.DefaultContainerInfoRequest()
	allContainers, err := client.AllDockerContainers(&request)
	if detectError(err) {
		return nil
	}

	return allContainers
}

func askMiddleware(server string) []byte {

	resp, err := http.Get(middleware + server)

	if detectError(err) == false {

		defer resp.Body.Close()
		content, err := ioutil.ReadAll(resp.Body)
		detectError(err)
		return content
	}

	return nil
}

func detectError(err error) bool {

	if err != nil {
		log.Println(err)
		return true
	} else {
		return false
	}
}

func main() {

	http.HandleFunc("/", handleRoot)

	http.HandleFunc("/view", handleView)

	http.HandleFunc("/redirect", handleRedirect)

	fmt.Println("Listening on http://127.0.0.1" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
