# mr-weasel

## Overview

Personal Multi-Tool Telegram bot written in Go, sqlx and official Telegram API.

- **Ping**: ping pong command for health-check
- **Car**: Browse and manage your cars to track fuel consumption and service expenses
- **Holiday**: Manage your business holidays
- **Other**: youtube to mp3, voice changer with python cli etc.

## Structure

- **commands**: Bot command handlers
- **tgclient**: Telegram api interactions
- **tgmanager**: Telegram bot event manager
- **storage**: Database access layer
- **migrations**: SQL migration scripts (Goose for versioning)

## Build

Follow these steps to build and run the project:

```sh
cp .env.example .env  # prepare the '.env' with default settings
make go-tools         # install necessary go tools into the 'build' folder
make build            # build the app
make run              # build and start the app
```

