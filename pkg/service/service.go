package service

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"smtp2communicator/pkg/logger"
	"text/template"
	"time"

	"go.uber.org/zap"
)

var serviceTemplate = `[Unit]
Description={{.Description}}
StartLimitIntervalSec=500
StartLimitBurst=5

[Service]
ExecStart={{.ExecStart}}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

var (
	serviceInstallationPath = "/etc/systemd/system/"
	log                     *zap.SugaredLogger
)

type FilesToCopy struct {
	Source      string
	Destination string
	Chmod       fs.FileMode
}

type Service struct {
	Name                    string
	Description             string
	ExecStart               string
	ServiceInstallationPath string
	FilesToCopy             []FilesToCopy
}

// AddFileToCopy add files to be installed/uninstalled
//
// This function is adding a file to FilesToCopy struct which will be
// installed to specified location on install and deleted on uninstall.
//
// Parameters:
//
// - source (string): path to file to copy
// - desctination (string): path where file should be copied to
// - chmod (fs.FileMode): permissions to be set on the desctination file
//
// Returns:
//
// - n/a
func (s *Service) AddFileToCopy(source string, desctination string, chmod fs.FileMode) {
	ftc := FilesToCopy{
		Source:      source,
		Destination: desctination,
		Chmod:       chmod,
	}

	s.FilesToCopy = append(s.FilesToCopy, ftc)
}

// getName returns Systemd service file name
//
// This function appends ".service" to name of the service as defined in Service struct.
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - name (string): service file name
func (s *Service) getName() string {
	return s.Name + ".service"
}

// getPath returns Systemd service file path
//
// This function concatenates serviceInstallationPath and return from getName
// which results in path where Systemd service file should be installed.
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - path (string): service file path
func (s *Service) getPath() (path string) {
	return serviceInstallationPath + s.getName()
}

// renderTemplate renders service file template and saves it to serviceFile
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) renderTemplate(serviceFile *os.File) (err error) {
	t := template.New("service")

	r, _ := t.Parse(serviceTemplate)
	err = r.Execute(serviceFile, s)
	if err != nil {
		return
	}
	return
}

// copyFiles processes and copies files as per configuration in FilesToCopy
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) copyFiles() (err error) {
	var srcHandle, dstHandle *os.File
	for _, f := range s.FilesToCopy {

		srcHandle, err = os.Open(f.Source)
		if err != nil {
			return
		}

		dstHandle, err = os.Create(f.Destination)
		if err != nil {
			return
		}

		_, err = io.Copy(dstHandle, srcHandle)
		if err != nil {
			log.Error(err)
			return
		}

		err = os.Chmod(f.Destination, f.Chmod)
		if err != nil {
			log.Error(err)
			return
		}
	}
	return
}

// Install installs Systemd service file
//
// This function copies all defined files to its defined locations, creates
// service file and reloads Systemd daemon.
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) Install(ctx context.Context) (err error) {
	err = s.copyFiles()
	if err != nil {
		return
	}

	serviceFile, err := os.OpenFile(s.getPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer serviceFile.Close()

	err = s.renderTemplate(serviceFile)
	if err != nil {
		return
	}

	err = s.DaemonReload(ctx)

	return
}

// Uninstall uninstalls Systemd service file
//
// This function deletes all defined in installation files from its desctination locations
// deletes service file and reloads Systemd daemon.
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) Uninstall(ctx context.Context) (err error) {
	for _, f := range s.FilesToCopy {
		// remove service file
		err = os.Remove(f.Destination)
		if err != nil {
			return
		}
	}

	err = s.DaemonReload(ctx)

	return
}

// InstallEnableStart wraps Install, Enable and Start methods
//
// This is convenience method that automates installation, enabling and starting a service.
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) InstallEnableStart() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()

	err = s.Install(ctx)

	if err != nil {
		cancel()
	}
	err = s.Enable(ctx)
	if err != nil {
		cancel()
	}
	err = s.Start(ctx)
	if err != nil {
		cancel()
	}
	return
}

// StopDisableUninstall wraps Stop, Disable and Uninstall methods
//
// This is convenience method that automates removal of a service.
//
// Parameters:
//
// - n/a
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) StopDisableUninstall() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.Stop(ctx)
	if err != nil {
		cancel()
	}

	err = s.Disable(ctx)
	if err != nil {
		cancel()
	}

	err = s.Uninstall(ctx)
	if err != nil {
		cancel()
	}
	return
}

// Start starts a Systemd service
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) Start(ctx context.Context) (err error) {
	err = execute(ctx, "systemctl", "start", s.getName())
	return err
}

// Stop stops a Systemd service
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) Stop(ctx context.Context) (err error) {
	err = execute(ctx, "systemctl", "stop", s.getName())
	return err
}

// Enable enables a Systemd service
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) Enable(ctx context.Context) (err error) {
	err = execute(ctx, "systemctl", "enable", s.getName())
	return err
}

// Disable disables a Systemd service
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) Disable(ctx context.Context) (err error) {
	err = execute(ctx, "systemctl", "disable", s.getName())
	return err
}

// DaemonReload reloads a Systemd service
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - err (error): error if any or nil
func (s *Service) DaemonReload(ctx context.Context) (err error) {
	err = execute(ctx, "systemctl", "daemon-reload")
	return err
}

// execute executes Systemd commands used by Service struct
//
// # This is where actual Systemd commands are executed
//
// Parameters:
//
// - ctx (context.Context): context
// - command (string): shell command to be executed
// - cmdArgs (..string): any number of arguments for the command
//
// Returns:
//
// - err (error): error if any or nil
func execute(ctx context.Context, command string, cmdArgs ...string) (err error) {
	if ctx.Err() != nil {
		log.Debug("Cancelling...")
		return errors.New("Context cancelled")
	}

	log.Debugf("executing command %s(%v)", command, cmdArgs)

	cmd := exec.CommandContext(ctx, command, cmdArgs...)

	var output []byte
	output, err = cmd.CombinedOutput()
	log.Debugf("command error: %v, output: %+v", err, string(output))
	return
}

// New creates a new Service struct
//
// This function creates a new service struct.
//
// Parameters:
//
// - logger (*zap.SugaredLogger): logger to be used in this Service struct
//
// Returns:
//
// - service (Service): a new Service struct
func New(ctx context.Context) (service Service) {
	log = logger.LoggerFromContext(ctx)
	return Service{
		ServiceInstallationPath: "/etc/systemd/system/",
	}
}
