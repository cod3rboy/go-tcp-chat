package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type ChatClient struct {
	conn   *net.TCPConn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewChatClient(conn *net.TCPConn) *ChatClient {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &ChatClient{
		conn:   conn,
		reader: reader,
		writer: writer,
	}
}

func (c *ChatClient) Close() {
	c.conn.Close()
}

func (c *ChatClient) SendMessage(msg string) error {
	_, err := c.writer.WriteString(fmt.Sprintf("%s%c", msg, '\000'))
	if err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *ChatClient) GetMessage() (string, error) {
	msg, err := c.reader.ReadString('\000')
	msg = strings.TrimRight(msg, "\000")
	return msg, err
}
