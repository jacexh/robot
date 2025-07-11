package main

import (
	"net/http"

	"github.com/jacexh/robot"
)

func main() {
	camera := robot.NewCamera()
	if err := camera.Start(); err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/video", camera.Stream)
	if err := http.ListenAndServe(":9999", mux); err != nil {
		panic(err)
	}
}
