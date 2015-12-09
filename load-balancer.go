package main

import (
	"github.com/google/cadvisor/client"
	info "github.com/google/cadvisor/info/v1"

	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

var filter string = "dummy"
var monitor string = "http://cadvisor:8080"
var port string = ":7500"

// determine the container with the more available RAM
func getLessLoaded(cadvisor string) string {

	allContainers := getAllContainerInfo(cadvisor)

	var alias, ret string
	var kbFree uint64
	tabContainers := make(map[string]uint64)

	for _, item := range allContainers {
		alias = item.Aliases[0]
		kbFree = item.Stats[0].Memory.Usage / 1024

		tabContainers[alias] = kbFree
	}
	ret = alias // for later comparison

	for key, _ := range tabContainers {

		// verifying only wanted containers
		if strings.Contains(key, filter) {
			if tabContainers[key] < tabContainers[ret] {
				ret = key
			}
		}
	}

	//fmt.Printf("\nLess loaded : %s with %d kb of RAM used\n", ret, tabContainers[ret])
	return ret
}

func handleView(w http.ResponseWriter, r *http.Request) {
	log.Println("GET vue")
	w.Write([]byte(getLessLoaded(monitor)))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("GET racine")
	w.Write([]byte("Hello World !"))
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	log.Println("GET redirige")

	lessLoaded := getLessLoaded(monitor)
	address, err := net.LookupHost(lessLoaded)

	if err != nil {
		log.Print("No address...")
		return
	}

	log.Println(lessLoaded)
	log.Println(address)

	lessLoaded = "http://" + address[0]

	http.Redirect(w, r, lessLoaded, 307)
}

func getAllContainerInfo(cadvisor string) []info.ContainerInfo {

	client, err := client.NewClient(cadvisor)
	printError(err)

	request := info.DefaultContainerInfoRequest()
	allContainers, err := client.AllDockerContainers(&request)
	printError(err)

	return allContainers
}

func printError(err error) {

	if err != nil {
		log.Println(err)
	}
}

func main() {

	server := http.NewServeMux()

	rootHandler := http.HandlerFunc(handleRoot)
	server.Handle("/", rootHandler)

	viewHandler := http.HandlerFunc(handleView)
	server.Handle("/view", viewHandler)

	redirectHandler := http.HandlerFunc(handleRedirect)
	server.Handle("/redirect", redirectHandler)

	fmt.Println("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, server)
}
