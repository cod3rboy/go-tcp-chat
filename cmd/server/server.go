package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/cod3rboy/go-tcp-chat/common"
)

const (
	port = 4000
)

type App struct {
	listener         net.Listener
	ErrorLog         *log.Logger
	InfoLog          *log.Logger
	UsernameChan     chan UsernameValidation
	AddClientChan    chan *Client
	RemoveClientChan chan *Client
	MessageChan      chan common.Message
}

func createServer() (*App, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	errorLog := log.New(os.Stdout, "ERROR", log.Ldate|log.Ltime|log.Llongfile)
	infoLog := log.New(os.Stdout, "INFO", log.Ldate|log.Ltime)

	return &App{
		listener:         listener,
		ErrorLog:         errorLog,
		InfoLog:          infoLog,
		UsernameChan:     make(chan UsernameValidation),
		AddClientChan:    make(chan *Client),
		RemoveClientChan: make(chan *Client),
		MessageChan:      make(chan common.Message),
	}, nil
}

func (app *App) Serve() {
	// Set chat manager
	go app.clientManager()
	// Wait for each incoming client connection
	for {
		// accept the connection
		conn, err := app.listener.Accept()
		if err != nil {
			app.ErrorLog.Println("incoming connection error: ", err)
			continue
		}
		// convert to TCP connection
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			app.ErrorLog.Println("failed to obtain TCP connection: ", err)
			continue
		}
		// serve connection on new goroutine
		go app.connectionHandler(tcpConn)
	}
}

func (app *App) connectionHandler(conn *net.TCPConn) {
	client, err := app.handshake(conn)
	if err != nil {
		app.ErrorLog.Println("client handshake failed:", err)
		return
	}
	app.InfoLog.Println("new client connected:", client.ID)
	app.AddClientChan <- client
	defer client.Close()
	for {
		msg, err := client.GetMessage()
		if err != nil {
			app.ErrorLog.Println(err)
			break
		}
		app.MessageChan <- common.Message{
			User:      client.ID,
			Timestamp: time.Now(),
			Body:      msg,
		}
	}
	app.InfoLog.Println("client", client.ID, "has disconnected")
	app.RemoveClientChan <- client
}

func (app *App) handshake(conn *net.TCPConn) (*Client, error) {
	client := NewClient(conn)
	if err := client.SendMessage("Welcome to the TCP Chat Server"); err != nil {
		return nil, err
	}
	for {
		username, err := client.GetMessage()
		if err != nil {
			return nil, err
		}
		validation := UsernameValidation{
			Username:  username,
			ValidChan: make(chan bool),
		}
		app.UsernameChan <- validation
		isValid := <-validation.ValidChan
		if isValid {
			client.ID = username
			if err := client.SendMessage("OK"); err != nil {
				return nil, err
			}
			// handshake complete
			break
		}
		if err := client.SendMessage("FAIL"); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func (app *App) clientManager() {
	clients := make(map[string]*Client)
	for {
		select {
		case client := <-app.AddClientChan:
			clients[client.ID] = client

		case client := <-app.RemoveClientChan:
			delete(clients, client.ID)

		case validation := <-app.UsernameChan:
			_, usernameExists := clients[validation.Username]
			validation.ValidChan <- !usernameExists

		case msg := <-app.MessageChan:
			for _, client := range clients {
				client.SendMessage(msg.Encode())
			}
		}
	}
}
