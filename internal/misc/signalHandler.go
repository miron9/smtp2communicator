package misc

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"smtp2communicator/pkg/logger"
)

// signalHandler handles system signals interrupting this tool
//
// This function catches system calls terminating this tool and cleans up if needed before exiting.
//
// Parameters:
//
// - ctx (context.Context): context
// - cronSendmailMTAPath (string): path where cron will look for sendmail
// - mtaStubInstalled (bool): flag indicating if we linked ourselves to sendmail
//
// Returns:
//
// - n/a
func SignalHandler(ctx context.Context, cronSendmailMTAPath string, mtaStubInstalled bool) {
	log := logger.LoggerFromContext(ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		switch <-signals {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP:
			log.Info("Termination requested, staring clean up...")
			err := SendmailMTAUninstall(ctx, false, cronSendmailMTAPath, mtaStubInstalled)
			log.Info("Exiting...")
			if err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}()
}
