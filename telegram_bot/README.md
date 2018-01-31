# Telegram Bot
> Bot used for operational tasks at RHBRLAB

This is a simple Telegram Bot, written in Golang, that interacts with Ansible Tower through RESTful APIs. The idea is to have a bot that can executes Ansible Playbook by simple Telegram commands.

## Binary Build

To make sure your executable is 100% static, you should build with:

```bash
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' rhbrlab-bot.go
```

## Container Build

To make things easier to run and maintein, you should use it as a container. To build a Docker container image, just use the `Dockerfile` included on this repository, and run:

```bash
docker build . -t rhbrlab-bot
```

## Execution

The bot requires some environment variables to run. You should run it with:

```bash
docker run -e TELEGRAM_TOKEN="<Telegram Bot API Token>" \
           -e TELEGRAM_CHATID="<Telegram Chat ID>" \
           -e TOWER_URL="<Tower API URL>" \
           -e TOWER_USER="<Tower User>" \
           -e TOWER_PASS="<Tower Password>" \
           -e TOWER_START_TEMPLATE_ID="<Job Template ID>" \
           -e TOWER_STOP_TEMPLATE_ID="<Job Template ID>"
           -itd rhbrlab-bot
```