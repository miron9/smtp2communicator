package stdin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"smtp2communicator/internal/common"
	"smtp2communicator/pkg/logger"

	"go.uber.org/zap"
)

func TestProcessStdin(t *testing.T) {
	ctx := context.Background()
	l, _ := zap.NewDevelopment()
	log := l.Sugar()
	ctx = logger.ContextWithLogger(ctx, log)

	// define a test message
	testMsg := common.Message{
		From:    "root (Cron Daemon)",
		To:      "user",
		Subject: "Cron <user@desktop> df -h",
		Body: `Filesystem Size Used Avail Use% Mounted on
tmpfs 6.2G 2.7M 6.2G 1% /run
/dev/nvme0n1p3 94G 47G 43G 52% /
tmpfs 31G 5.0M 31G 1% /dev/shm
tmpfs 5.0M 16K 5.0M 1% /run/lock
tmpfs 128M 0 128M 0% /mnt/ramdisk
tmpfs 31G 0 31G 0% /run/qem`,
	}

	// define the test message as a single string
	message := fmt.Sprintf("From: %s\n"+
		"To: %s\n"+
		"Subject: %s\n"+
		"MIME-Version: 1.0\n"+
		"Content-Type: text/plain; charset=UTF-8\n"+
		"Content-Transfer-Encoding: 8bit\n"+
		"X-Cron-Env: <SHELL=/bin/sh>\n"+
		"X-Cron-Env: <HOME=/home/user>\n"+
		"X-Cron-Env: <LOGNAME=user>\n\n%s", testMsg.From, testMsg.To, testMsg.Subject, testMsg.Body)

	msgChan := make(chan common.Message, 10) // Adjust the buffer size as needed

	data := strings.NewReader(message)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// start the function to to be tested
	go ProcessStdin(ctx, data, msgChan, &wg, 1)

	// allow some time to process
	time.Sleep(100 * time.Millisecond)

	// indicate the ProcessStdin doesn't have to wait any more
	wg.Done()

	// receive and verify the result
	for msg := range msgChan {
		if msg.Subject != testMsg.Subject {
			t.Fatalf("Received SUBJECT is not matching expected one: '%s' != '%s'", testMsg.Subject, msg.Subject)
		}

		if msg.From != testMsg.From {
			t.Fatalf("Received FROM is not matching expected one: '%s' != '%s'", testMsg.From, msg.From)
		}

		if msg.To != testMsg.To {
			t.Fatalf("Received TO is not matching expected one: '%s' != '%s'", testMsg.To, msg.To)
		}

		if msg.Body != testMsg.Body {
			t.Fatalf("Received BODY is not matching expected one: '%s' != '%s'", testMsg.Body, msg.Body)
		}
	}
}
