package telegram

import (
	"fmt"
	"strings"

	"smtp2communicator/internal/common"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"go.uber.org/zap"
)

// SendTelegramMsg sends a message to Telegram communicator
//
// This function takes 2 arguments, 'conf' being Telegram specific configuration
// as defined in accompanying configuration file and the actual message to be sent 'msg'.
//
// Parameters:
//
// - conf (TelegramChannel): Telegram configuration struct
// - msg (string): message to be sent
//
// Returns:
// - err (error): if any or nil
func SendTelegramMsg(log *zap.SugaredLogger, conf common.TelegramChannel, newMessage common.Message) (err error) {
	b, err := gotgbot.NewBot(conf.BotKey, nil)
	if err != nil {
		log.Errorf("Error creating new bot: %v", err)
		return err
	}

	msgFmtd := formatTelegramMessage(log, newMessage)
	// Telegram can take up to 4096 long message with all formating included
	chunkedMsgs := common.Splitter(4050, msgFmtd)
	totalMsgs := len(chunkedMsgs)
	msgCount := 1
	for chunkId, chunk := range chunkedMsgs {
		if err != nil {
			return err
		}
		chunk = fmt.Sprintf("(%d/%d)\n%s", msgCount, totalMsgs, chunk)
		_, err = b.SendMessage(conf.UserId, markdownMessage(chunk), &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})
		if err != nil {
			log.Errorf("Error sending Telegram message %d: %v", chunkId, err)
			return err
		}
		msgCount++
	}

	log.Infof("Telegram message sent")
	return nil
}

// formatTelegramMessage formats message to Telegram communicator
//
// This function parses the 'message' struct and formats a message that is to be sent to the Telegram
//
// Parameters:
//
// - msg (message): the 'message' struct
//
// Returns:
//
// - msgFmtd (string): a formatted message as a code block
func formatTelegramMessage(log *zap.SugaredLogger, msg common.Message) (msgFmtd string) {
	msgFmtd = fmt.Sprintf("Time: %s\nFrom: %s\nTo: %s\nSubject: %s\n\n%s",
		msg.Time, msg.From, msg.To, msg.Subject, msg.Body)

	replacer := strings.NewReplacer(
		"{", "\\{",
		"}", "\\}",
		"-", "\\-",
		".", "\\.",
		"+", "\\+",
		"=", "\\=",
	)

	replacer.Replace(msgFmtd)

	return msgFmtd
}

func markdownMessage(message string) (msgFmtd string) {
	msgFmtd = fmt.Sprintf("```\n%s\n```", message)
	return msgFmtd
}
