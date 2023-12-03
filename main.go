package main

import (
	"log"
	"net"
)

type MessageType int

type Message struct {
	Conn net.Conn
	Type MessageType
	Text []byte
}

const Port = "9595"

const (
	NewClient MessageType = iota
	DisconnectedClient
	NewMessage
)

func main() {
	ln := startServer()

	channel := make(chan Message)
	go Chat(channel)

	conn := awaitNewConnection(ln, channel)
	go HandleConnection(conn, channel)
}

func HandleConnection(conn net.Conn, channel chan Message) {
	for {
		buffer := make([]byte, 512)
		_, err := conn.Read(buffer)
		if err != nil {
			log.Println("Client Disconnected. Closing the connection.")
			channel <- Message{Conn: conn, Type: DisconnectedClient}
			break
		}
		channel <- Message{Conn: conn, Type: NewMessage, Text: buffer}
	}
}

func Chat(broadcastChan chan Message) {
	clients := make(map[string]net.Conn)
	for {
		msg := <-broadcastChan
		if msg.Type == NewClient {
			clients[msg.Conn.RemoteAddr().String()] = msg.Conn
		} else if msg.Type == DisconnectedClient {
			delete(clients, msg.Conn.RemoteAddr().String())
			msg.Conn.Close()
		} else if msg.Type == NewMessage {
			for _, conn := range clients {
				if conn.RemoteAddr().String() == msg.Conn.RemoteAddr().String() {
					continue
				}
				conn.Write(msg.Text)
			}
		} else {
			log.Println("Received an unknown message type = ", msg.Type)
		}
	}
}

func startServer() net.Listener {
	ln, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen on port %s. Shutting down ...\n", Port)
	}

	log.Printf("Listening on port %s\n", Port)
	return ln
}

func awaitNewConnection(ln net.Listener, channel chan Message) net.Conn {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Could not accept connection. ", err)
		}

		channel <- Message{Conn: conn, Type: NewClient}
		log.Println("connection accepted. ", conn.RemoteAddr())

		return conn
	}
}
