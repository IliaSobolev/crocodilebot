# crocodilebot

This is a telegram bot game crocodile. In this game, one user guesses a word and the rest must guess the word.

--------------------------------
- - [Installation](#installation)
- - [How to use](#how\to\user)
--------------------------------
## Installation
The program only works with [docker](https://www.docker.com/) and [redis](https://redis.io/docs/stack/get-started/install/docker/).
After that you need to clone repository from github:

    git clone https://github.com/doginwatermelon/crocodilebot

Then go to the directory:

    cd crocodilebot/

Also you need to generate a token for Telegram Bot API in [BotFather](https://t.me/botfather) and create an .env file and specify there:

    TOKEN = "your token"

and start project:

    go run main.go

## How to use
Add your bot to the group and write /start
