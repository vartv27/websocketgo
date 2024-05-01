package main

import (
    "fmt"
    "log"
    "net/http"
    "net"
    "strconv"
   // "time"
    "github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // Connected clients
var users = make(map[*websocket.Conn]string) //
var broadcast = make(chan Message)           // Broadcast channel

// Define our Message struct
type Message struct {
    Username string `json:"username"`
    Text string `json:"text"`
    X int `json:"x"`
    Y int `json:"y"`
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading to WebSocket:", err)
        return
    }
    defer conn.Close()

    // Your WebSocket handling logic here
}

var i int = 0
func handleConnections(w http.ResponseWriter, r *http.Request) {
    // Upgrade initial GET request to a websocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    // Make sure we close the connection when the function returns
    defer ws.Close()

    // Register our new client
    clients[ws] = true
    users[ws] = "U" + strconv.Itoa(i)
    i++

    for {
        var msg Message
        // Read in a new message as JSON and map it to a Message object
        err := ws.ReadJSON(&msg)
        msg.Username = users[ws]
        if err != nil {
            log.Printf("error: %v", err)
            delete(clients, ws)
            delete(users, ws)
            break
        }
	   // fmt.Println(msg) 
        // Send the newly received message to the broadcast channel
        broadcast <- msg
    }
}

func handleMessages() {
    for msg := range broadcast {
        // Grab the next message from the broadcast channel
       // msg := <-broadcast
        // Send it out to every client that is currently connected
        for client := range clients {
			//msg.x = 45;
	     //	msg.Username = "U1"
			//msg.Username = users[client].username
		//	fmt.Println(msg)
			//time.Sleep(8 * time.Second) 
            err := client.WriteJSON(msg)

            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}

func main() {
    fmt.Println(GetOutboundIP())
    
    // Create a simple file server
    fs := http.FileServer(http.Dir("public"))
    http.Handle("/", fs)

    http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	    fmt.Println("test_in")
		fmt.Fprintln(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Test</title>
		</head>
		<body>
		     test socket
		</body>
		</html>
		`)	
	})

    // Configure websocket route
    http.HandleFunc("/ws", handleConnections)

    // Start listening for incoming chat messages
    go handleMessages()

    // Start the server on localhost port 8000 and log any errors
    fmt.Println("Server running on :3000")
    err := http.ListenAndServe(":3000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
    //log.Fatal(app.listening)
}


