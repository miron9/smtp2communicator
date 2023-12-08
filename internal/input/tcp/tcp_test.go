package internal

import (
	//"bufio"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"smtp2communicator/internal/common"
	"smtp2communicator/pkg/logger"

	"go.uber.org/zap"
)

func TestProcessTCP(t *testing.T) {
	// Create a context for the test
	ctx := context.Background()
	l, _ := zap.NewDevelopment()
	log := l.Sugar()
	ctx = logger.ContextWithLogger(ctx, log)

	// Create a channel for receiving messages
	msgChan := make(chan common.Message, 10) // Adjust the buffer size as needed

	// Choose an available port for testing
	testPort := 12345

	// Start the ProcessTCP function in a goroutine
	go ProcessTCP(ctx, msgChan, testPort)

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Create a test TCP connection to the server
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", testPort))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	// define a test message
	testMsg := common.Message{
		From:    "cron@example.com",
		To:      "user@example.com",
		Subject: "hello test",
		Body:    "body body body",
	}

	// reader := bufio.NewReader(conn)

	// send HELO, FROM and TO headers and read responses
	conn.Write([]byte("HELO example.com\r\n"))
	// reader.ReadString('\n')

	conn.Write([]byte(fmt.Sprintf("MAIL FROM: %s\r\n", testMsg.From)))
	// reader.ReadString('\n')

	conn.Write([]byte(fmt.Sprintf("RCPT TO: %s\r\n", testMsg.To)))
	// reader.ReadString('\n')

	conn.Write([]byte("DATA\r\n"))
	// reader.ReadString('\n')

	conn.Write([]byte(fmt.Sprintf("Subject: %s\r\n", testMsg.Subject)))
	// reader.ReadString('\n')

	conn.Write([]byte("\r\n"))

	conn.Write([]byte(fmt.Sprintf("%s", testMsg.Body)))

	conn.Write([]byte("\r\n.\r\n"))
	// reader.ReadString('\n')

	conn.Write([]byte("QUIT\r\n"))
	// reader.ReadString('\n')
	conn.Close()

	time.Sleep(100 * time.Millisecond)

	// Close the msgChan to signal the server to stop processing
	close(msgChan)

	for msg := range msgChan {
		if msg.Subject != testMsg.Subject {
			t.Fatalf("Received SUBJECT is not matching expected one: '%s' != '%s'", testMsg.Subject, msg.Subject)
		}

		if msg.From != testMsg.From {
			t.Fatalf("Received FROM is not matching expected one: '%s' != '%s'", testMsg.From, msg.From)
		}

		if msg.To != testMsg.To {
			t.Fatalf("Received TO is not matching expected one: '%s' != '%s'", testMsg.To, msg.To)
		}

		if msg.Body != testMsg.Body {
			t.Fatalf("Received BODY is not matching expected one: '%s' != '%s'", testMsg.Body, msg.Body)
		}
	}
}
