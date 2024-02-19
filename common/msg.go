package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidFormat = errors.New("invalid message format")
)

type Message struct {
	User      string
	Body      string
	Timestamp time.Time
}

func (m Message) Encode() string {
	return fmt.Sprintf("%d\t%s\t%s", m.Timestamp.UnixMilli(), m.User, m.Body)
}

func DecodeMessage(message string) (msg Message, err error) {
	parts := strings.SplitN(message, "\t", 3)
	if len(parts) != 3 {
		err = ErrInvalidFormat
		return
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		err = ErrInvalidFormat
		return
	}

	msg.Timestamp = time.UnixMilli(ts)
	msg.User = parts[1]
	msg.Body = parts[2]

	return msg, nil
}
