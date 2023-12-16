package slack

import (
	"fmt"

	"smtp2communicator/internal/common"

	"go.uber.org/zap"

	"github.com/slack-go/slack"
)

func SendSlackMsg(log *zap.SugaredLogger, conf common.SlackChannel, newMessage common.Message) (err error) {
	s := slack.New(conf.BotKey)

	msgFmtd := formatMessage(log, newMessage)
	chunkedMsgs := common.Splitter(4050, msgFmtd)
	totalMsgs := len(chunkedMsgs)
	msgCount := 1
	for chunkId, chunk := range chunkedMsgs {
		if err != nil {
			return err
		}
		chunk = fmt.Sprintf("(%d/%d)\n%s", msgCount, totalMsgs, chunk)

		chunk = markdownMessage(chunk)

		_, _, _, err := s.SendMessage(conf.UserId, slack.MsgOptionText(chunk, true))
		if err != nil {
			log.Errorf("Error sending Slack message %d: %v", chunkId, err)
			return err
		}
		msgCount++
	}

	log.Infof("Slack message sent")
	return nil
}

// formatMessage formats message to Slack communicator
//
// This function parses the 'message' struct and formats a message that is to be sent to the Slack
//
// Parameters:
//
// - msg (message): the 'message' struct
//
// Returns:
//
// - msgFmtd (string): a formatted message as a code block
func formatMessage(log *zap.SugaredLogger, msg common.Message) (msgFmtd string) {
	msgFmtd = fmt.Sprintf("Time: %s\nFrom: %s\nTo: %s\nSubject: %s\n\n%s",
		msg.Time, msg.From, msg.To, msg.Subject, msg.Body)

	return msgFmtd
}

func markdownMessage(message string) (msgFmtd string) {
	msgFmtd = fmt.Sprintf("```\n%s\n```", message)
	return msgFmtd
}
