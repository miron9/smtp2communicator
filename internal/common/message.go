package common

import "time"

type Message struct {
	Time    time.Time
	Headers map[string]string
	From    string
	To      string
	Subject string
	Body    string
}
