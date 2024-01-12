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
//
// Returns:
//
// - err (error): error if any or nil
func SendmailMTAInstall(ctx context.Context, cronSendmailMTAPath string) (err error) {
	log := logger.LoggerFromContext(ctx)

	// Check if we're root user
	if os.Geteuid() != 0 {
		log.Error("Sendmail stub un/install must be run as the root user. Skipping.")
		return errors.New("Not a root user")
	}

	_, err = os.Stat(cronSendmailMTAPath)
	if errors.Is(err, fs.ErrNotExist) {
		myExecPath := os.Args[0]
		if os.Symlink(myExecPath, cronSendmailMTAPath) != nil {
			log.Errorf("Can't symlink Sendmail: %v", err)
		}
		log.Infof("Stub symlinked to '%s'", cronSendmailMTAPath)
	} else {
		log.Warn("Some other MTA already present, not installing the stub")
		return errors.New("Sendmail already linked")
	}

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

	// Check if we're root user
	if os.Geteuid() != 0 {
		log.Error("Sendmail stub un/install must be run as the root user. Skipping.")
		return errors.New("Not a root user")
	}

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
// - exit (bool): flag indicating if the programme should terminate after this function is done
// - err (error): error if any or nil
func MtaOnly(ctx context.Context, installMTAOnly *bool, uninstallMTAOnly *bool, cronSendmailMTAPath string, mtaStubInstalled bool) (exit bool, err error) {
	// MTA sendmail
	if !*installMTAOnly || !*uninstallMTAOnly {
		// link as an MTA only
		if *installMTAOnly {
			exit = true
			err = SendmailMTAInstall(ctx, cronSendmailMTAPath)
			if err == nil {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
		// unlink as an MTA
		if *uninstallMTAOnly {
			exit = true
			err = SendmailMTAUninstall(ctx, true, cronSendmailMTAPath, mtaStubInstalled)
			if err == nil {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
	} else {
		exit = true
		err = errors.New("The 'installMTA' and 'uninstallMTA' flags can't be used together!")
	}
	return
}
