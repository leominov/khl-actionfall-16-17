package main

import (
	"flag"
	"net/http"

	"github.com/Sirupsen/logrus"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	hub := newHub()
	go hub.run()
	go hub.readPump()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	logrus.Fatal(http.ListenAndServe(*addr, nil))
}
