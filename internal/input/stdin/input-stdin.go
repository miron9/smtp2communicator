package stdin

import (
	"bufio"
	"io"
	"net/mail"
	"strings"
	"time"

	c "smtp2communicator/internal/common"

	"github.com/DusanKasan/parsemail"
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
// - input (io.Reader): where to read from the input (usually os.Stdin)
// - msgProcessed (chan<- bool): exit status indicating if we did any work here
// - msgChan (chan<- c.Message): channel to send a message to
//
// Returns:
//
// - argName (argType): [description]
// - err (error): error if any or nil
func readStdin(log *zap.SugaredLogger, input io.Reader, msgProcessed chan<- bool, msgChan chan<- c.Message) {
	scanner := bufio.NewScanner(input)

	body := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		body = append(body, line)
	}

	bodyText := strings.Join(body, "\n")

	parsedMsg, err := parsemail.Parse(strings.NewReader(bodyText))
	if err != nil {
		log.Error(err)
		return
	}

	// send info that there was no message in the body hence not sending anything and return
	if len(parsedMsg.TextBody) == 0 {
		msgProcessed <- false
		return
	}

	// If Year is 0 or 1 int then we replace that date with current time
	var msgTime time.Time
	if parsedMsg.Date.Year() <= 1 {
		msgTime = time.Now()
	} else {
		msgTime = parsedMsg.Date
	}

	newMessage := c.Message{
		Time: msgTime,
	}
	newMessage.From = getEmailAddr(parsedMsg.From, parsedMsg.Header["From"][0])
	newMessage.To = getEmailAddr(parsedMsg.To, parsedMsg.Header["To"][0])
	newMessage.Subject = parsedMsg.Subject
	newMessage.Body = parsedMsg.TextBody

	// send message to dispatcher
	msgChan <- newMessage
	close(msgChan)

	// indicate we've done work
	msgProcessed <- true
	close(msgProcessed)
}

// getEmailAddr extracts email address from parsed message
//
// This function is returning email address as it was specified in the source email
// but it will look for it first in From attribute and if not present then it will be
// read from headers. This is addressing issue where for example Cron can set sender and
// recipient to be invalid email addresses and in such case the parsemail module
// used here is not going to set it in From attribute but raw value is still present in headers.
//
// Parameters:
//
// - parsedEmailList ([]*mail.Address): content of the From or To attributes from parsed email
// - unParsedEmail (string): this is basically defult value to be returned if parsedEmailList is empty
//
// Returns:
//
// - fmtdField (string): comma seprated email addresses
func getEmailAddr(parsedEmailList []*mail.Address, unParsedEmail string) (fmtdField string) {
	if len(parsedEmailList) > 0 {
		return mailToString(parsedEmailList)
	} else {
		return unParsedEmail
	}
}

func mailToString(emailList []*mail.Address) (fmtdField string) {
	for _, value := range emailList {
		fmtdField = value.String() + ", "
	}
	fmtdField = fmtdField[:len(fmtdField)-2]
	return fmtdField
}
