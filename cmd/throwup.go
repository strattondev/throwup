package main

import (
	"flag"
	"github.com/strattonw/throwup"
	"log"
	"net/http"
	"strconv"
)

func main() {
	endpointCount := flag.Int("count", 1, "Amount of endpoints to create")
	storage := flag.String("storage", "", "Directory to store uploaded files")
	successMessage := flag.String("success", "", "Message to display on upload success")
	singleUse := flag.Bool("single", true, "Disable endpoints once file uploaded")
	singleUseMessage := flag.String("singleMessage", "", "Message to display for pages that have been disabled")
	port := flag.Int("port", 8080, "Server port")

	flag.Parse()

	throwup.New(throwup.Config{
		*endpointCount,
		*storage,
		*successMessage,
		*singleUse,
		*singleUseMessage}).Run(http.DefaultServeMux)

	log.Printf("Starting server localhost:%d\n", *port)
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
