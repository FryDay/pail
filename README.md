# Pail

Pail is a simple Discord bot, heavily inspired by xkcd's [Bucket](https://github.com/zigdon/xkcd-Bucket)

Pail was thrown together in a few hours and as such is horribly written at the moment. Feel free to submit PRs with improvements to functionality, code quality, bug fixes, or anything else.

## Getting Started

### Dependencies

* Go
* sqlite3
* [reflex](https://github.com/cespare/reflex) to use hotreload

### Setup

1. Create pail config directory at `~/.config/pail/`
2. Copy `pail.toml` from the examples folder to new config directory
3. [Configure pail](#configuration)
4. Run pail once, either using `hotreload.sh` or `go run cmd/pail/main.go`. This will create an empty database. Quit pail using `ctrl+c`
5. Run initial database population using `cat examples/pail.sql | sqlite3 ~/.config/pail/pail.db`

### Configuration

* Token: Your bot's unique token
* ReplaceChance: The chance a phrase will be replaced using a `replace` action in the regex table. 5 = 5%, 100 = 100%, etc.
* RandomInterval: How often your bot will pick a random fact if there is no activity in minutes
* RandomChannels: List of channel IDs to send a random fact on. To get a channel's ID, right click on it in Discord and select `Copy ID`. Leaving the list empty will result in no random facts.

## License

This project is licensed under the MIT License - see the LICENSE.md file for details
