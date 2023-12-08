package internal

import (
	"context"
	"fmt"
	"net"
	"os"

	"smtp2communicator/internal/common"
	"smtp2communicator/pkg/logger"
)

// processTCP handles messages incoming via TCP
//
// This function starts TCP listener and handles incoming traffic by starting
// a new goroutine for each connection.
//
// Parameters:
//
// - ctx (context.Context): context
// - msgChan (chan message): channel to pass received messages for sending
// - port (int): port number to listen on
//
// Returns:
//
// - n/a
func ProcessTCP(ctx context.Context, msgChan chan<- common.Message, port int) {
	log := logger.LoggerFromContext(ctx)

	// determine hostname to be used
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "hostname-not-available"
		log.Errorf("Error, can't get hostname, using fake one: %s", hostname)
	}

	// Start the SMTP server on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error starting SMTP server: %v", err)
	}
	defer listener.Close()

	log.Infof("SMTP stub listening on port %d\n", port)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Warnf("Error accepting connection: %v", err)
			continue
		}

		// Handle each incoming connection in a separate goroutine
		go handleConnection(log, hostname, conn, msgChan)
	}
}
