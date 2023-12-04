package main

import (
	"log"
	"net"
	"time"
)

// TODO:
// rate limiter
// throttling

type MessageType int

type Message struct {
	Conn      net.Conn
	Type      MessageType
	Text      []byte
	CreatedAt time.Time
}

const Port = "9595"

const (
	NewClient MessageType = iota
	DisconnectedClient
	NewTextClient
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
			channel <- NewMessage(conn, DisconnectedClient, nil)
			break
		}
		channel <- NewMessage(conn, NewTextClient, buffer)
	}
}

func Chat(broadcastChan chan Message) {
	clients := make(map[string]net.Conn)
	for {
		msg := <-broadcastChan
		if msg.Type == NewClient {
			clients[msg.Conn.RemoteAddr().String()] = msg.Conn
			log.Println("New client = ", msg.Conn.RemoteAddr().String())
		} else if msg.Type == DisconnectedClient {
			delete(clients, msg.Conn.RemoteAddr().String())
			msg.Conn.Close()
			log.Println("Client disconnected. Connection closed.")
		} else if msg.Type == NewTextClient {
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

func NewMessage(conn net.Conn, msgType MessageType, buffer []byte) Message {
	if msgType == NewClient {
		return Message{Conn: conn, Type: NewClient, CreatedAt: time.Now()}
	} else if msgType == DisconnectedClient {
		return Message{Conn: conn, Type: DisconnectedClient, CreatedAt: time.Now()}
	} else if msgType == NewTextClient {
		return Message{Conn: conn, Type: NewTextClient, Text: buffer, CreatedAt: time.Now()}
	} else {
		return Message{Conn: conn, CreatedAt: time.Now()}
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

		channel <- NewMessage(conn, NewClient, nil)
		log.Println("connection accepted. ", conn.RemoteAddr())

		return conn
	}
}
