# smtp2communicator

A simple tool to replace local SMTP server and forward received mail to selected communicator (like Telegram)

## What it does?

This tool reads mail submitted on port 25 or via STDIN and forwards it to all configured channels (like the Telegram communicator).

## Motivation

I got tired of configuring mail forwarding or fighting Gmail to not refuse mail from my mail server or creating dedicated email account just to get my cronjob reports from my home server, etc.

## Limitations

### Telegram

Sends only one message trimmed to 4096 characters (if longer than that), text only.

## Configuration

This tool can be configured via YAML configuration file. To get example configuration file either check [configs](configs/sampleConfig.yaml) or run this tool like so:

```go
./smtp2communicator -configurationExample > smtp2communicator.yaml
```

The configuration file can be specified explicitly via `-configuration` parameter or it will be searched for in those location in following order:

- current working directory,
- location of this executable,
- users home directory,
- /etc/

Configuration file should be named `smtp2communicator.yaml`.

### Outputs

Also at the time of writing this supported outputs are:

- local file and
- Telegram.

### Telegram

Telegram configuration requires configured BOT with API key and message recipient's ID.

To create a BOT and get a token follow [this](https://core.telegram.org/bots/tutorial#obtain-your-bot-token) official documentation (assumes you're already Telegram user).

Obtaining user ID can be achieved by talking to "GetIDs Bot" - just type this name to Telegram search or click [here](https://web.telegram.org/k/#@getidsbot) (web Telegram). Once you open a chat with the bot just send to it "/start" to get some info about yourself (ID included).

Now just enter these to relevant places inside the smtp2communicator.yaml generated earlier.

## Tested on

So far this has been tested only on Ubuntu Linux 22 and 23
