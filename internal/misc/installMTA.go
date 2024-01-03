package misc

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"smtp2communicator/pkg/logger"
)

// SendmailMTAInstall links this tool to "sendmail" command
//
// This function links and thus installs this to be seen in the system
// as an MTA. The "sendmail" path to be linked is described here cronSendmailMTAPath.
//
// Parameters:
//
// - ctx (context.Context): context
// - cronSendmailMTAPath (string): path where cron will look for sendmail
// - mtaStubInstalled (bool): flag indicating if we linked ourselves to sendmail
//
// Returns:
//
// - err (error): error if any or nil
func SendmailMTAInstall(ctx context.Context, cronSendmailMTAPath string, mtaStubInstalled bool) (err error) {
	log := logger.LoggerFromContext(ctx)

	_, err = os.Stat(cronSendmailMTAPath)
	if errors.Is(err, fs.ErrNotExist) {
		myExecPath := os.Args[0]
		if os.Symlink(myExecPath, cronSendmailMTAPath) != nil {
			log.Errorf("Can't symlink Sendmail: %v", err)
		}
		log.Infof("Stub symlinked to '%s'", cronSendmailMTAPath)
		mtaStubInstalled = true
	} else {
		log.Warn("Some other MTA already present, not installing the stub")
		return errors.New("Sendmail already linked")
	}

	signalHandler(ctx, cronSendmailMTAPath, mtaStubInstalled)

	return
}

// SendmailMTAUninstall unlinks this tool from "sendmail" command
//
// This function unlinks and thus uninstalls this tool from being seen in the system
// as an MTA. The "sendmail" path to be unlinked is described here cronSendmailMTAPath
//
// Parameters:
//
// - ctx (context.Context): context
// - uninstallMTAOnly (*bool): flag indicating MTA should be uninsalled
// - cronSendmailMTAPath (string): path where cron will look for sendmail
// - mtaStubInstalled (bool): flag indicating if we linked ourselves to sendmail
//
// Returns:
//
// - err (error): error if any or nil
func SendmailMTAUninstall(ctx context.Context, uninstallMTAOnly bool, cronSendmailMTAPath string, mtaStubInstalled bool) (err error) {
	log := logger.LoggerFromContext(ctx)

	if mtaStubInstalled || uninstallMTAOnly {
		err = os.Remove(cronSendmailMTAPath)
		if err != nil {
			log.Errorf("Can't unlink this stub from Sendmail ('%s'): %v", cronSendmailMTAPath, err)
			return
		}
		log.Infof("Sendmail unlinked from: %s", cronSendmailMTAPath)
	}
	return
}

// signalHandler handles system signals interrupting this tool
//
// This function catches system calls terminating this tool and cleans up if needed before exiting.
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - n/a
func signalHandler(ctx context.Context, cronSendmailMTAPath string, mtaStubInstalled bool) {
	log := logger.LoggerFromContext(ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		switch <-signals {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP:
			log.Info("Termination requested, staring clean up...")
			SendmailMTAUninstall(ctx, false, cronSendmailMTAPath, mtaStubInstalled)
			log.Info("Exiting...")
			os.Exit(1)
		}
	}()
}

// mtaOnly installs this tool as an MTA
//
// This function is linking this tool to sendmail command (if not already present).
//
// Parameters:
//
// - ctx (context.Context): context
// - installMTAOnly (*bool): flag indicating MTA should be insalled
// - uninstallMTAOnly (*bool): flag indicating MTA should be uninsalled
// - cronSendmailMTAPath (string): path where cron will look for sendmail
// - mtaStubInstalled (bool): flag indicating if we linked ourselves to sendmail
//
// Returns:
//
// - n/a
func MtaOnly(ctx context.Context, installMTAOnly *bool, uninstallMTAOnly *bool, cronSendmailMTAPath string, mtaStubInstalled bool) {
	log := logger.LoggerFromContext(ctx)

	// MTA sendmail
	if !*installMTAOnly || !*uninstallMTAOnly {
		// link as an MTA only
		if *installMTAOnly {
			if SendmailMTAInstall(ctx, cronSendmailMTAPath, mtaStubInstalled) == nil {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
		// unlink as an MTA
		if *uninstallMTAOnly {
			if SendmailMTAUninstall(ctx, true, cronSendmailMTAPath, mtaStubInstalled) == nil {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
	} else {
		log.Error("The 'installMTA' and 'uninstallMTA' flags can't be used together!")
		os.Exit(1)
	}
}
