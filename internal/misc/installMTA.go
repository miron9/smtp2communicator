package misc

import (
	"context"
	"errors"
	"io/fs"
	"os"

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

	return nil // resets potential err returned by os.Stat
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
		}
		// unlink as an MTA
		if *uninstallMTAOnly {
			exit = true
			err = SendmailMTAUninstall(ctx, true, cronSendmailMTAPath, mtaStubInstalled)
		}
	} else {
		exit = true
		err = errors.New("The 'installMTA' and 'uninstallMTA' flags can't be used together!")
	}
	return
}
