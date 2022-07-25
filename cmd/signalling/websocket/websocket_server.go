package websocket

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}
	defer c.Close()
	for {
	}
}

func createSignallingWs() {
	http.HandleFunc("/echo", echo)
	http.ListenAndServe("0.0.0.0:8088", nil)
}
