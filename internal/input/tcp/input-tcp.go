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

	"github.com/DusanKasan/parsemail"
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

	newBody := []string{}

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

				newBody = append(newBody, fmt.Sprintf("From: %s\n", result[1]))

				conn.Write([]byte("250 OK (mail from)\n"))
			}
		}

		if utils.MatchString("start", line, "rcpt to") {
			if result, err := utils.KvExtractor(line); err == nil {
				newMessage.To = result[1]

				newBody = append(newBody, fmt.Sprintf("To: %s\n", result[1]))

				conn.Write([]byte("250 OK (rcpt to)\n"))
			}
		}

		if utils.MatchString("exact", line, "data") {
			conn.Write([]byte("354 OK (data)\n"))
			break
		}
	}

	// here we read actual message
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Error("Can't read from stdin")
		}

		// a single '.' on it's own means end of message
		if utils.MatchString("exact", line, ".") {
			conn.Write([]byte("250 OK body\n"))
			continue

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

		newBody = append(newBody, line)
	}

	msgString := strings.Join(newBody, "")
	msgStringLen := len(msgString)
	msgString = msgString[:msgStringLen-2] // remove trailing \r\n
	parsedMsg, err := parsemail.Parse(strings.NewReader(msgString))
	if err != nil {
		log.Error(err)
		return
	}

	if len(parsedMsg.TextBody) == 0 {
		return
	}

	newMessage.From = parsedMsg.From[0].String()
	newMessage.To = parsedMsg.To[0].String()
	newMessage.Subject = parsedMsg.Subject
	newMessage.Body = parsedMsg.TextBody

	msgChan <- newMessage
}
