package misc

import (
	"context"
	"sync"

	"smtp2communicator/internal/common"
	"smtp2communicator/internal/output/file"
	"smtp2communicator/internal/output/slack"
	"smtp2communicator/internal/output/telegram"
	"smtp2communicator/pkg/logger"
)

// dispatcher is a siple function that calls channels passing them received
// message for sending to its destination
//
// Parameters:
//
// - ctx (context.Context): context
// - dstChanConf (common.Channels): configuration for all channels as specified in configuration yaml
// - msgChan (<-chan message): message struct channel
// - wg (sync.WaitGroup): channel to pass received messages to
//
// Returns:
//
// - n/a
func Dispatcher(ctx context.Context, dstChanConf common.Channels, msgChan <-chan common.Message, wg *sync.WaitGroup) {
	log := logger.LoggerFromContext(ctx)

	log.Info("dispatcher started")
	for incomingMsg := range msgChan {
		log.Debugf("got message with subject: %s", incomingMsg.Subject)
		telegram.SendTelegramMsg(log, dstChanConf.Telegram, incomingMsg)
		slack.SendSlackMsg(log, dstChanConf.Slack, incomingMsg)
		file.SaveEmailToFile(log, dstChanConf.File, incomingMsg)
	}
	log.Debug("Channel with incoming messages closed")

	// indicate we're done here so no need to wait any more;
	// this will be executed only when msg channel is closed
	wg.Done()
}
