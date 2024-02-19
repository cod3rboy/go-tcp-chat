package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cod3rboy/go-tcp-chat/common"
)

const (
	port = 4000
)

type App struct {
	listener       *net.TCPListener
	ErrorLog       *log.Logger
	InfoLog        *log.Logger
	ChatManager    *ChatManager
	wg             sync.WaitGroup
	connectionChan chan *net.TCPConn
	doneChan       chan bool
}

func NewChatServer() (*App, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	tcpListener, _ := listener.(*net.TCPListener)

	errorLog := log.New(os.Stdout, "ERROR", log.Ldate|log.Ltime|log.Llongfile)
	infoLog := log.New(os.Stdout, "INFO", log.Ldate|log.Ltime)

	return &App{
		listener:       tcpListener,
		ErrorLog:       errorLog,
		InfoLog:        infoLog,
		ChatManager:    NewChatManager(),
		doneChan:       make(chan bool),
		connectionChan: make(chan *net.TCPConn),
	}, nil
}

func (app *App) GoServe() {
	app.wg.Add(1)
	go func() {
		app.ChatManager.Start()

		defer func() {
			close(app.doneChan)
			close(app.connectionChan)
			app.wg.Done()
		}()

		app.wg.Add(1)
		// wait for connections
		go func() {
			defer app.wg.Done()
			for {
				tcpConn, err := app.listener.AcceptTCP()
				if err != nil {
					app.ErrorLog.Println("error accepting connection: ", err)
					return
				}
				app.connectionChan <- tcpConn
			}
		}()

		connections := make(map[*net.TCPConn]bool, 0)
		handlersWg := sync.WaitGroup{}
		for {
			select {
			case connection := <-app.connectionChan:
				// handle new connection
				handlersWg.Add(1)
				connections[connection] = true
				go func() { // handler goroutine
					defer handlersWg.Done()
					app.handleConnection(connection)
					delete(connections, connection)
				}()
			case <-app.doneChan:
				app.listener.Close()
				for connection := range connections {
					// this will also finish active handler goroutines
					connection.Close()
				}
				handlersWg.Wait()
				app.ChatManager.Stop()
				return
			}
		}
	}()
}

func (app *App) Stop() {
	app.doneChan <- true
	app.wg.Wait()
}

func (app *App) handleConnection(conn *net.TCPConn) {
	participant, err := app.handshake(conn)
	if err != nil {
		app.ErrorLog.Println("client handshake failed:", err)
		return
	}

	app.InfoLog.Println("new client connected:", participant.ID)
	app.ChatManager.AddParticipant(participant)
	defer participant.Close()

	for {
		msg, err := participant.GetMessage()
		if err != nil {
			app.ErrorLog.Println(err)
			break
		}
		app.ChatManager.BroadcastMessage(common.Message{
			User:      participant.ID,
			Timestamp: time.Now(),
			Body:      msg,
		})
	}

	app.InfoLog.Println("participant ", participant.ID, "has disconnected")
	app.ChatManager.RemoveParticipant(participant)
}

func (app *App) handshake(conn *net.TCPConn) (*Participant, error) {
	participant := NewParticipant(conn)
	if err := participant.SendMessage("Welcome to the TCP Chat Server"); err != nil {
		return nil, err
	}
	for {
		username, err := participant.GetMessage()
		if err != nil {
			return nil, err
		}
		if !app.ChatManager.UsernameExists(username) {
			participant.ID = username
			if err := participant.SendMessage("OK"); err != nil {
				return nil, err
			}
			// handshake complete
			break
		}
		if err := participant.SendMessage("FAIL"); err != nil {
			return nil, err
		}
	}
	return participant, nil
}
