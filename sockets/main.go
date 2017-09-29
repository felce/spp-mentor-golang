package main

import (
	"log"
	"net/http"
	"os"

	"github.com/googollee/go-socket.io"
)

func main() {

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		so.BroadcastTo("cursor field", "get info")

		so.Join("cursor field")
		var idCursor, idName, userColor string

		so.On("user", func(user, color, nameId, cursorId string) {
			idName = nameId
			idCursor = cursorId
			userColor = color
			if user != "" {
				so.BroadcastTo("cursor field", "append user", user, color, nameId)
			}
		})

		so.On("coords", func(coordsX, coordsY string) {

			//circle client cursor
			so.Emit("new coords", coordsX, coordsY, userColor, idCursor)
			so.BroadcastTo("cursor field", "new coords", coordsX, coordsY, userColor, idCursor)
		})

		so.On("disconnection", func() {
			so.BroadcastTo("cursor field", "close", idCursor, idName)
			log.Println("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
