# task — a tiny coding-task backlog

A single-user Go CLI that stores tasks in `~/.config/task/tasks.json` (override with `TASK_FILE`). It has no third-party dependencies.

## Install

```sh
go install github.com/goelab-inc/tasks-cli@latest
```

The binary is named after the module (`tasks-cli`). To get a `task` command, alias it or build with `go build -o task .`.

## Usage

```sh
task add fix login bug !high @api +2h due:fri
task add "write regression tests" !m @auth +30m due:tomorrow
task                  # no arguments: interactive quick-add prompt
task ls               # backlog, most important first
task ls @api
task ls --all         # include completed tasks
task catchup          # one task to do now plus two runner-ups
task next             # alias for catchup
task start ab123
task done ab123
task edit ab123 revise login tests !high @api +1h due:2026-12-01
task bump ab123
task snooze ab123 +2h
task show ab123
task rm ab123
```

Quick-add accepts `!h`/`!high`/`!3`, `!m`/`!2`, `!l`/`!1`; a project `@name`; an estimate such as `+30m` or `+2h`; and `due:today`, `due:tomorrow`, `due:weekday` (the next such weekday, including when today matches), or `due:YYYY-MM-DD`. Other words form the task description. IDs are short and can be used as unique prefixes. `ls` ranks the backlog; `--all` also includes completed tasks. `catchup` omits completed and currently snoozed items.

## Catch-up scoring

Higher scores rise to the top. The score is the sum of: priority (low 10, medium 20, high 30); due urgency (overdue +50, otherwise linearly declining from +35 today to 0 at seven days); aging (+0.5 per backlog day, capped at +14); in-progress bonus (+15); and quick-win bonus (+8 for estimates up to 30 minutes, +4 up to two hours). This intentionally favors urgent work and tasks already underway, while giving neglected items a gentle lift.

## Development

```sh
go test ./...
go vet ./...
```

Tasks are stored as JSON. Set `NO_COLOR` to disable ANSI colors; color is also suppressed when stdout is not a terminal.
