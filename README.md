# missionctl

A command-line client for [Mission Control](https://missioncontrol.dev) projects and time
entries, for people and agents.

## Install

```bash
brew install kmarsh/tap/missionctl
```

Or download a build for macOS, Linux, or Windows from
[Releases](https://github.com/kmarsh/missionctl/releases), or `go install
github.com/kmarsh/missionctl@latest`.

## Setup

Generate an API key at [missioncontrol.dev/api](https://missioncontrol.dev/api), then either
export it:

```bash
export MISSIONCTL_API_KEY=mc_...
```

or put it in `~/.mission_control.toml`, which the Mission Control Mac app reads too:

```toml
api_key = "mc_..."
```

## Usage

### Projects

```bash
# Enabled projects (add --all or --disabled for the rest)
missionctl projects

# One project, by name or id
missionctl projects show "Globex Portal"
```

### Time entries

```bash
# This week's entries, from Monday
missionctl time

# Today's, a date range, or one project's
missionctl time --today
missionctl time --from 2026-09-01 --to 2026-09-30
missionctl time -p Acme

# Log time; the date defaults to today
missionctl time log "Fixed the login form" -p Acme -d 1:30
missionctl time log "Standup" -p "Globex Portal" -d 15 --date yesterday
missionctl time log "Pro bono review" -p Acme -d 45m --no-billable

# Change an entry; only the options given change
missionctl time edit ENTRY_ID -d "2 hrs" -m "Pairing on the login form"

# Show or delete an entry (delete asks first; --yes skips that)
missionctl time show ENTRY_ID
missionctl time delete ENTRY_ID
```

### Scripting

```bash
# Hours per project this month
missionctl time --from "$(date +%Y-%m-01)" --json \
  | jq -r 'group_by(.project.name)[] | "\(.[0].project.name): \(map(.minutes) | add / 60) h"'
```

In a terminal it prints tables. Piped, or with `--json`, it prints the API's JSON (an array for
lists), and errors go to stderr as JSON, so agents and scripts get structured output by default.

- **Projects** can be named instead of given by id: an exact name ignoring case, or a unique
  partial name. Names match enabled projects only, except with `projects show --all`.
- **Dates** are `YYYY-MM-DD`, `today`, or `yesterday`.
- **Durations** can be `1:30`, `:45`, a number of minutes (`45`), or hours and minutes with
  units: `45m`, `10 min`, `1.5h`, `2 hrs`, `1h30m`, `"1 hour 20 min"`.
- **Exit codes:** 0 success, 1 API or network error, 64 bad usage or input, 77 API key rejected,
  78 no API key or unreadable config.

## Releasing

Push a `v*` tag. GitHub Actions runs [GoReleaser](https://goreleaser.com), which builds macOS,
Linux, and Windows binaries, publishes the release, and updates the cask in
[kmarsh/homebrew-tap](https://github.com/kmarsh/homebrew-tap).

```bash
git tag -a v0.3.0 -m "missionctl 0.3.0" && git push origin v0.3.0
```

The workflow needs a `HOMEBREW_TAP_GITHUB_TOKEN` secret: a fine-grained token with contents
write access to kmarsh/homebrew-tap.

## License

[MIT](LICENSE)
