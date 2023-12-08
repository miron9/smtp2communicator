package telegram

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"go.uber.org/zap"
	"smtp2communicator/internal/common"
)

// sendTelegramMsg sends a message to Telegram communicator
//
// This function takes 2 arguments, 'conf' being Telegram specific configuration
// as defined in accompanying configuration file and the actual message to be sent 'msg'.
//
// Parameters:
//
// - conf (TelegramChannel):  Telegram configuration struct
// - msg (string): message to be sent
//
// Returns:
// - err (error): if any or nil
func SendTelegramMsg(log *zap.SugaredLogger, conf common.TelegramChannel, newMessage common.Message) error {

	msg, err := formatTelegramMessage(log, newMessage)
	if err != nil {
		return err
	}

	b, err := gotgbot.NewBot(conf.BotKey, nil)
	if err != nil {
		log.Errorf("Error creating new bot: %v", err)
		return err
	}

	_, err = b.SendMessage(conf.UserId, msg, &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
	if err != nil {
		log.Errorf("Error sending Telegram message: %v", err)
		return err
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
// - err (error): error if any or nil
func formatTelegramMessage(log *zap.SugaredLogger, msg common.Message) (msgFmtd string, err error) {
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

	msgFmtd = fmt.Sprintf("```\n%s\n```", msgFmtd)

	if len(msgFmtd) > 4096 {
		log.Error("Message not sent as too long")
		//return "", errors.New("Message to long")
		msgFmtd = fmt.Sprintf("%s...", msgFmtd[:4093])
	}

	return msgFmtd, nil
}
