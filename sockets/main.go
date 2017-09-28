package main

import (
	"log"
	"net/http"

	"github.com/googollee/go-socket.io"
)

func main() {

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")

		so.Join("cursor field")
		i := ""

		so.On("coords", func(coordsX, coordsY, color, id string) {
			i = id
			//circle client cursor
			//so.Emit("new coords", coordsX, coordsY, color, id)
			so.BroadcastTo("cursor field", "new coords", coordsX, coordsY, color, id)
		})

		so.On("disconnection", func() {
			so.BroadcastTo("cursor field", "close", i)
			log.Println("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
