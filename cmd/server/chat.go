package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/cod3rboy/go-tcp-chat/common"
)

type UsernameInfo struct {
	ID        string
	ValidChan chan bool
}

type Participant struct {
	ID     string
	reader bufio.Reader
	writer bufio.Writer
	conn   *net.TCPConn
}

func NewParticipant(conn *net.TCPConn) *Participant {
	return &Participant{
		conn:   conn,
		reader: *bufio.NewReader(conn),
		writer: *bufio.NewWriter(conn),
	}
}

func (c *Participant) Close() {
	c.conn.Close()
}

func (c *Participant) GetMessage() (string, error) {
	msg, err := c.reader.ReadString('\000')
	msg = strings.TrimRight(msg, "\000")
	return msg, err
}

func (c *Participant) SendMessage(msg string) error {
	_, err := c.writer.WriteString(fmt.Sprintf("%s%c", msg, '\000'))
	if err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

type ChatManager struct {
	participants         map[string]*Participant
	validateUsernameChan chan UsernameInfo
	connectChan          chan *Participant
	disconnectChan       chan *Participant
	messageChan          chan common.Message
	doneChan             chan bool
	wg                   sync.WaitGroup
}

func NewChatManager() *ChatManager {
	return &ChatManager{
		participants:         make(map[string]*Participant),
		validateUsernameChan: make(chan UsernameInfo),
		connectChan:          make(chan *Participant),
		disconnectChan:       make(chan *Participant),
		messageChan:          make(chan common.Message),
		doneChan:             make(chan bool),
	}
}

func (c *ChatManager) Start() {
	c.wg.Add(1)
	go func() {
		defer func() {
			close(c.validateUsernameChan)
			close(c.connectChan)
			close(c.disconnectChan)
			close(c.messageChan)
			close(c.doneChan)
			c.wg.Done()
		}()
		for {
			select {
			case participant := <-c.connectChan:
				// save new participant
				c.participants[participant.ID] = participant

			case participant := <-c.disconnectChan:
				delete(c.participants, participant.ID)

			case usernameInfo := <-c.validateUsernameChan:
				_, usernameExists := c.participants[usernameInfo.ID]
				usernameInfo.ValidChan <- usernameExists

			case msg := <-c.messageChan:
				for _, client := range c.participants {
					client.SendMessage(msg.Encode())
				}
			case <-c.doneChan:
				// close all participants connections
				for _, participant := range c.participants {
					participant.Close()
				}
				return
			}
		}
	}()
}

func (c *ChatManager) Stop() {
	c.doneChan <- true
	c.wg.Wait()
}

func (c *ChatManager) UsernameExists(username string) bool {
	usernameInfo := UsernameInfo{
		ID:        username,
		ValidChan: make(chan bool),
	}
	c.validateUsernameChan <- usernameInfo
	exists := <-usernameInfo.ValidChan
	close(usernameInfo.ValidChan)
	return exists
}

func (c *ChatManager) AddParticipant(participant *Participant) {
	c.connectChan <- participant
}

func (c *ChatManager) RemoveParticipant(participant *Participant) {
	c.disconnectChan <- participant
}

func (c *ChatManager) BroadcastMessage(message common.Message) {
	c.messageChan <- message
}
