package stdin

import (
	"bufio"
	"io"
	"strings"
	"time"

	c "smtp2communicator/internal/common"

	"smtp2communicator/pkg/utils"

	"go.uber.org/zap"
)

// readStdin reads message from standard input
//
// This function reads a message received on stdin into c.Message structs and
// sends it to dispacher via msgChan channel.
//
// Parameters:
//
// - log (*zap.SugaredLogger): logger
// - msgProcessed (chan<- bool): exit status indicating if we did any work here
// - msgChan (chan<- c.Message): channel to send a message to
//
// Returns:
//
// - argName (argType): [description]
// - err (error): error if any or nil
func readStdin(log *zap.SugaredLogger, input io.Reader, msgProcessed chan<- bool, msgChan chan<- c.Message) {
	scanner := bufio.NewScanner(input)
	// scanner := bufio.NewScanner(os.Stdin)

	newMessage := c.Message{
		Time: time.Now(),
	}

	headers := map[string]string{}
	headersCron := []string{}

	body := []string{}

	for scanner.Scan() {
		line := scanner.Text()

		if result, err := utils.KvExtractor(line); err == nil {
			if utils.MatchString("exact", result[0], "x-cron-env") {
				headersCron = append(headersCron, result[1])
			} else {
				headers[strings.ToLower(result[0])] = result[1]
			}
		}

		// an empty line in this block will mean begin of actual message body
		if utils.MatchString("exact", line, "") {
			break
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		body = append(body, line)
	}

	// send info that no message in body hence not sending anything and exit
	if len(body) == 0 {
		msgProcessed <- false
		return
	}

	newMessage.From = headers["from"]
	newMessage.To = headers["to"]
	newMessage.Subject = headers["subject"]
	newMessage.Body = strings.Join(body, "\n")

	// send message to dispatcher
	msgChan <- newMessage
	close(msgChan)

	// indicate we've done work
	msgProcessed <- true
	close(msgProcessed)
}
