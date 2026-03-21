# gokillport

Find and kill processes listening on given ports. One command, no noise.

## Install

### Quick install (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/nopedal/gokillport/main/install.sh | sh
```

This downloads the right binary for your system and installs both `gokillport` and `killport` to `/usr/local/bin`.

### Go

```bash
go install github.com/nopedal/gokillport/cmd/gokillport@latest
go install github.com/nopedal/gokillport/cmd/killport@latest
```

### Homebrew (coming soon)

```bash
brew tap nopedal/tap
brew install gokillport
```

### From source

```bash
git clone https://github.com/nopedal/gokillport.git
cd gokillport
make install
```

## Usage

Both `gokillport` and `killport` work identically — use whichever you prefer.

Kill a process on a port:

```bash
killport 3000
```

Kill multiple ports at once:

```bash
killport 3000 8080 5432
```

Force kill (SIGKILL):

```bash
killport -f 3000
```

Dry run — see what would be killed:

```bash
killport -n 3000
```

List all listening processes:

```bash
killport -l
```

JSON output for scripting:

```bash
killport --json 3000
```

Quiet mode (errors only):

```bash
killport -q 3000
```

Custom timeout before escalation to SIGKILL:

```bash
killport -t 10 3000
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Send SIGKILL instead of SIGTERM |
| `--dry-run` | `-n` | Show what would be killed |
| `--list` | `-l` | List all listening processes |
| `--timeout` | `-t` | Seconds to wait before SIGKILL (default: 5) |
| `--quiet` | `-q` | Suppress output except errors |
| `--json` | | Output as JSON |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Process not found |
| 2 | Permission denied |
| 3 | Timeout |

## Cross-platform

Works on macOS, Linux, and Windows. Platform detection is automatic at runtime.

- **macOS**: uses `lsof`
- **Linux**: uses `ss` with `/proc/net/tcp` fallback
- **Windows**: uses `netstat -ano` and `taskkill`

## Releasing

Tag and push — GitHub Actions builds and uploads binaries automatically:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Or manually: `make release` (requires `gh` CLI).

## Build

```bash
make build          # current platform (both binaries)
make cross          # all platforms → dist/
make install        # install to $GOPATH/bin
make test           # run tests
```

## License

MIT
