package cmd

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
	mux.HandleFunc("/", camera.Stream)
	if err := http.ListenAndServe(":9999", mux); err != nil {
		panic(err)
	}
}
