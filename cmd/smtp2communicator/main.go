package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"os"
	"sync"

	c "smtp2communicator/internal/common"
	stdin "smtp2communicator/internal/input/stdin"
	tcp "smtp2communicator/internal/input/tcp"
	m "smtp2communicator/internal/misc"
	"smtp2communicator/pkg/logger"
	"smtp2communicator/pkg/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	stdinTimeout             = 2
	cronSendmailMTAPath      = "/usr/sbin/sendmail" // this is path at which crontab checks if MTA (sendmail) is installed
	binInstallPath           = "/usr/local/bin/"
	configurationInstallPath = "/etc/"
	projectName              = "smtp2communicator"
)

var (
	logLevel          zapcore.Level
	log               *zap.SugaredLogger
	conf              *c.Configuration
	mtaStubInstalled  bool
	configurationPath string // effective configuration file path
	thisBinaryPath    string
)

func init() {
	// init some var
	// mtaStubInstalled indicates if we linked ourselves to sendmail
	// this is used to unlink on termination
	mtaStubInstalled = false

	// init logger
	l, err := zap.NewDevelopment()
	// l, err := zap.NewProduction()
	if err != nil {
		panic("Can't create logger")
	}

	log = l.Sugar()
	defer l.Sync()

	thisBinaryPath, err = utils.GetMyPath()
	if err != nil {
		log.Error(err)
		panic("Can't get my own path")
	}
}

func main() {
	var err error
	var defaultConfigurationPath string

	// set-up context
	ctx := context.Context(context.Background())

	// put logger into context
	ctx = logger.ContextWithLogger(ctx, log)

	// define default value for -- configuration flag
	defaultConfigurationPath, err = c.FindConfigurationFile(ctx, thisBinaryPath, projectName+".yaml")

	var configurationNotFound error
	if err != nil {
		configurationNotFound = err
	}

	// flags
	// with shorthand
	configurationFileFlag := new(string)
	flag.StringVar(configurationFileFlag, "configuration", defaultConfigurationPath, "path to configuration file")
	flag.StringVar(configurationFileFlag, "c", defaultConfigurationPath, "path to configuration file")
	verbosityLevelFlag := new(string)
	flag.StringVar(verbosityLevelFlag, "verbosity", "info", "logging level, one of debug, info, error")
	flag.StringVar(verbosityLevelFlag, "v", "info", "logging level, one of debug, info, error")

	// without shorthand
	installMTAOnlyFlag := flag.Bool("installMTA", false, "link this tool to 'sendmail' making it effectively being seen as an MTA")
	uninstallMTAOnlyFlag := flag.Bool("uninstallMTA", false, "unlink this tool from 'sendmail'")
	systemdInstallFlag := flag.Bool("systemdInstall", false, "create Systemd service, enable and start it")
	systemdUninstallFlag := flag.Bool("systemdUninstall", false, "stop, disable and delete Systemd service")
	configurationExample := flag.Bool("configurationExample", false, "print to stdout example configuration file")
	// TODO add option to allow to pass free text to the tool so any message (not only email) can be sent

	// sendmail flags, to support the way Cron invokes it to pipe a message to it via stdin
	// Nov 25 19:14:01 desktop cron[108918]: [/usr/sbin/sendmail -FCronDaemon -i -B8BITMIME -oem auser]
	flag.Bool("FCronDaemon", false, "does nothing, required by Cron")
	flag.Bool("i", false, "does nothing, required by Cron")
	flag.Bool("B8BITMIME", false, "does nothing, required by Cron")
	flag.Bool("oem", false, "does nothing, required by Cron")

	flag.Parse()

	// set logging level
	switch *verbosityLevelFlag {
	case "error":
		logLevel = zapcore.ErrorLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "debug":
		logLevel = zapcore.DebugLevel
	}
	log = log.WithOptions(zap.IncreaseLevel(logLevel))

	// put logger into context
	ctx = logger.ContextWithLogger(ctx, log)

	// -----------------------------------------------------

	// print example configuration to stdout and exit
	if *configurationExample {
		m.ConfigurationExample()
		os.Exit(0)
	}

	// load configuration from file
	conf = &c.Configuration{}
	err = conf.GetConfiguration(ctx, *configurationFileFlag)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			log.Error(err)
		} else {
			log.Errorf("The configuration file must be specified with --configuration flag or exist in one of the following locations: %s", configurationNotFound)
		}
		os.Exit(1)
	}

	// This will un/install this tools as a Systemd service and exit if either
	// of the flags has been defined
	if m.SystemdService(
		ctx,
		systemdInstallFlag,
		systemdUninstallFlag,
		configurationFileFlag,
		configurationInstallPath,
		thisBinaryPath,
		binInstallPath,
		projectName,
	) {
		os.Exit(0)
	}

	// This will un/install this tool as an MTA end exit if either of the flags
	// has been defined
	m.MtaOnly(ctx, installMTAOnlyFlag, uninstallMTAOnlyFlag, cronSendmailMTAPath, mtaStubInstalled)

	// channel that input sources pass received messages to dispatcher for sending
	msgChan := make(chan c.Message, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go m.Dispatcher(ctx, conf.Channels, msgChan, &wg)

	// process stdin input if any (exits if there was a message on stdin)
	if stdin.ProcessStdin(ctx, os.Stdin, msgChan, &wg, stdinTimeout) {
		os.Exit(0)
	}

	// stub set-up to allow Cron to send emails, this will be in place only
	// for the time of execution of this tool (as in opposition to
	// installMTAOnlyFlag, uninstallMTAOnlyFlag) and only if not other tool already linked to sendmail
	m.SendmailMTAInstall(ctx, cronSendmailMTAPath, mtaStubInstalled)

	// start listener and handle tcp connections
	tcp.ProcessTCP(ctx, msgChan, conf.Port)
}
