package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/GeistInDerSH/clearscreen"
	"github.com/cod3rboy/go-tcp-chat/common"
)

func main() {
	conn, err := net.Dial("tcp", ":4000")
	if err != nil {
		log.Fatalln("failed to connect with chat server: ", err)
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Fatalln("failed to convert into TCP connection: ", err)
	}

	client := NewChatClient(tcpConn)
	defer client.Close()

	welcomeMsg, err := client.GetMessage()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(welcomeMsg)
	reader := bufio.NewReader(os.Stdin)
	username := ""
	for {
		fmt.Print("Enter username: ")
		line, _ := reader.ReadString('\n')
		username = strings.TrimSpace(line)
		if err := client.SendMessage(username); err != nil {
			log.Println(err)
			return
		}
		msg, err := client.GetMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		if msg == "OK" {
			break
		}
		fmt.Println("choose different username!")
	}

	go func() {
		messages := make([]common.Message, 0)
		for {
			clearscreen.ClearScreen()
			for _, message := range messages {
				fmt.Printf("[%s]: %s\n", message.User, message.Body)
			}
			fmt.Printf("[%s] Enter Message: ", username)

			// wait for incoming message
			msgEncoded, err := client.GetMessage()
			if err != nil {
				return
			}
			message, err := common.DecodeMessage(msgEncoded)
			if err != nil {
				return
			}
			messages = append(messages, message)
		}
	}()

	for {
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if err := client.SendMessage(msg); err != nil {
			break
		}
	}
}
