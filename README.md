# convit

In an effort to make it easier to write conventional commit messages, `convit` is a command-line tool that allows you to write conventional commit messages.

## Install

```bash
# Install in the current directory
curl -sSL https://raw.githubusercontent.com/segersniels/convit/master/scripts/install.sh | bash
# Install in /usr/local/bin
curl -sSL https://raw.githubusercontent.com/segersniels/convit/master/scripts/install.sh | sudo bash -s /usr/local/bin
```

### Manual

1. Download the latest binary from the [releases](https://github.com/segersniels/convit/releases/latest) page for your system
2. Rename the binary to `convit`
3. Copy the binary to a location in your `$PATH`

## Usage

```
NAME:
   convit - Write conventional commit messages

USAGE:
   convit [global options] command [command options]

VERSION:
   0.1.0

COMMANDS:
   commit    Write a commit message
   generate  Write a commit message with the help of OpenAI
   config    Configure the app
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Generate

Experimental feature that uses OpenAI to assist with writing a conventional commit message. It looks at the currently staged changes that you want to commit and a user specified commit message to determine the type & optional scope of the commit.

```bash
convit generate
```

> This feature is _bring-your-own-key_ and requires the `OPENAI_API_KEY` environment variable to be set.
