package internal

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	c "smtp2communicator/internal/common"
	"smtp2communicator/pkg/utils"

	"go.uber.org/zap"
)

// handleConnection handles incoming TCP connection
//
// This function read a message frim the TCP connection received into c.Message
// struct and sends it to dispacher via msgChan channel.
//
// Parameters:
//
// - log (*zap.SugaredLogger): logger
// - hostname (string): hostname to use in welcome message
// - conn (net.Conn): tcp connection to read from
// - msgChan (chan<- c.Message): channel to send a message to
//
// Returns:
//
// - n/a
func handleConnection(log *zap.SugaredLogger, hostname string, conn net.Conn, msgChan chan<- c.Message) {
	defer conn.Close()

	newMessage := c.Message{
		Time: time.Now(),
	}

	connectionMessage := fmt.Sprintf("220 %s\n", hostname)
	conn.Write([]byte(connectionMessage))

	// Use a bufio.Reader to read lines from the connection
	reader := bufio.NewReader(conn)

	body := []string{}

	// read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Warnf("Can't read data")
		}

		if utils.MatchString("start", line, "ehlo") {
			conn.Write([]byte("250 OK welcome\n"))
		}
		if utils.MatchString("start", line, "helo") {
			conn.Write([]byte("250 OK welcome\n"))
		}

		if utils.MatchString("start", line, "mail from") {
			if result, err := utils.KvExtractor(line); err == nil {
				newMessage.From = result[1]
				conn.Write([]byte("250 OK (mail from)\n"))
			}
		}

		if utils.MatchString("start", line, "rcpt to") {
			if result, err := utils.KvExtractor(line); err == nil {
				newMessage.To = result[1]
				conn.Write([]byte("250 OK (rcpt to)\n"))
			}
		}

		if utils.MatchString("exact", line, "data") {
			conn.Write([]byte("354 OK (data)\n"))
			break
		}
	}

	// here we read actual message
	inMessageBodyBlock := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error("Can't read from stdin")
		}

		// if the line is subject then read it and skip adding to body
		if utils.MatchString("start", line, "subject") {
			if result, err := utils.KvExtractor(line); err == nil {
				newMessage.Subject = result[1]
			}
		}

		// an empty line in this block will mean begin of actual message body
		if utils.MatchString("exact", line, "") {
			inMessageBodyBlock = true
		}

		// get all lines of the message's body into the slice
		if inMessageBodyBlock {
			body = append(body, line)
		}

		// a single '.' on it's own means end of message
		if utils.MatchString("exact", line, ".") {
			conn.Write([]byte("250 OK body\n"))
			inMessageBodyBlock = false
		}

		if err != nil {
			if err == io.EOF {
				break // Connection closed by the client
			}
			log.Errorf("Error reading data: %v", err)
			return
		}

		if utils.MatchString("exact", line, "quit") {
			conn.Write([]byte("221 OK quit\n"))
			break
		}
	}
	// trim first and last line as it will contain word "data" and "."
	newMessage.Body = strings.Join(body[1:len(body)-1], "")

	// trim last carriage return and new line characters that were sent
	newMessage.Body = strings.TrimRight(newMessage.Body, "\r\n")

	if len(body) == 0 {
		return
	}

	msgChan <- newMessage
}
