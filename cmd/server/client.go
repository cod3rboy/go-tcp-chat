package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type UsernameValidation struct {
	Username  string
	ValidChan chan bool
}

type Client struct {
	ID     string
	reader bufio.Reader
	writer bufio.Writer
	conn   *net.TCPConn
}

func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		conn:   conn,
		reader: *bufio.NewReader(conn),
		writer: *bufio.NewWriter(conn),
	}
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) GetMessage() (string, error) {
	msg, err := c.reader.ReadString('\000')
	msg = strings.TrimRight(msg, "\000")
	return msg, err
}

func (c *Client) SendMessage(msg string) error {
	_, err := c.writer.WriteString(fmt.Sprintf("%s%c", msg, '\000'))
	if err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}
