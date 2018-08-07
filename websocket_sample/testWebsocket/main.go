package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) //connected clients
var broadcast = make(chan Message)           //broadcast channel
//config the upgrader
var upgrader = websocket.Upgrader{}

//define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	//create a simple file server
	fs := http.FileServer(http.Dir("../testWebsocketSource"))
	http.Handle("/", fs)
	//config websocket route
	http.HandleFunc("/ws", handleConnections)
	//start listening for incoming chat messages
	go handleMessages()
	//start the server on localhost 800 and log any error
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	//upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	//make sure we close the connection when the function returns
	defer ws.Close()
	//Register our new clinet
	clients[ws] = true
	for {
		var msg Message //read in a new message as JSON and map it to a message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		//send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		//grab the next message from the broadcast channel
		msg := <-broadcast
		//send it out to every client that is connect currently
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
