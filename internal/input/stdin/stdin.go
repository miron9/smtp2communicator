package stdin

import (
	"context"
	"io"
	"sync"
	"time"

	"smtp2communicator/pkg/logger"

	c "smtp2communicator/internal/common"
)

// processStdin handles messages incoming via STDIN
//
// This function is awaiting a message sent to this tool via standard input.
// If message is received it will be processed and the tool exits. If no
// message is received in defined time window then the STDIN channel times out
// and too continues to processTCP.
//
// Parameters:
//
// - ctx (context.Context): context
// - input (io.Reader): source of input message to read from, normally os.Stdin
// - msgChan (chan message): channel to pass received messages to
// - wg (sync.WaitGroup): channel to pass received messages to
// - stdinTimeout (int): seconds to wait for input on stdin
//
// Returns:
//
// - exit (bool): true if stdin message was processed and we should exit
func ProcessStdin(ctx context.Context, input io.Reader, msgChan chan c.Message, wg *sync.WaitGroup, stdinTimeout int) (exit bool) {
	log := logger.LoggerFromContext(ctx)

	// listen for a message on stdin first
	// if input present then prccess it and exit
	// or timeout and proceed to TCP listening
	log.Info("Stdin enabled")
	msgProcessed := make(chan bool, 1)
	go readStdin(log, input, msgProcessed, msgChan)

	select {
	case stdinResult := <-msgProcessed:
		log.Info("Stdin processing completed")
		if stdinResult {
			wg.Wait()
			log.Info("Message processed, exiting")
			return true
		} else {
			log.Info("No message processed")
		}
	case <-time.After(time.Duration(stdinTimeout) * time.Second):
		log.Infof("Stdin timed out after %ds", stdinTimeout)
	}
	return
}
