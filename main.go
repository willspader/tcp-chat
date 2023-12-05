package main

import (
	"log"
	"net"
	"time"
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
	NewTextClient
)

func main() {
	ln := startServer()

	channel := make(chan Message)
	go Chat(channel)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Could not accept connection. ", err)
		}

		channel <- NewMessage(conn, NewClient, nil)
		log.Println("connection accepted. ", conn.RemoteAddr())

		go HandleConnection(conn, channel)
	}
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
	lastNewTextClient := make(map[string]time.Time)
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
			lastTime := lastNewTextClient[msg.Conn.RemoteAddr().String()]
			if !lastTime.IsZero() && lastTime.After(time.Now().Add(-time.Second*5)) {
				msg.Conn.Write([]byte("The time elapse between messages is 5 seconds."))
			} else {
				lastNewTextClient[msg.Conn.RemoteAddr().String()] = time.Now()
				for _, conn := range clients {
					if conn.RemoteAddr().String() == msg.Conn.RemoteAddr().String() {
						continue
					}
					conn.Write(msg.Text)
				}
			}
		} else {
			log.Println("Unknown message type received = ", msg.Type)
		}
	}
}

func NewMessage(conn net.Conn, msgType MessageType, buffer []byte) Message {
	if msgType == NewClient {
		return Message{Conn: conn, Type: NewClient}
	} else if msgType == DisconnectedClient {
		return Message{Conn: conn, Type: DisconnectedClient}
	} else if msgType == NewTextClient {
		return Message{Conn: conn, Type: NewTextClient, Text: buffer}
	} else {
		return Message{Conn: conn}
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
