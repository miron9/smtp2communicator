package misc

import (
	"fmt"

	c "smtp2communicator/internal/common"

	"gopkg.in/yaml.v3"
)

// ConfigurationExample prints to stdout an example confiuration
func ConfigurationExample() {
	config := c.Configuration{
		Host: "127.0.0.1",
		Port: 25,
		Channels: c.Channels{
			File: c.FileChannel{
				Enabled: true,
				DirPath: "/path/to/files",
			},
			Telegram: c.TelegramChannel{
				Enabled: true,
				UserId:  123456789,
				BotKey:  "your_telegram_bot_api_key",
			},
			Slack: c.SlackChannel{
				Enabled: false,
			},
			Teams: c.TeamsChannel{
				Enabled: false,
			},
			Whatsup: c.WhatsupChannel{
				Enabled: false,
			},
		},
	}

	yConfig, err := yaml.Marshal(config)
	if err != nil {
		fmt.Print("can't print example")
	}

	fmt.Println(string(yConfig))
}
