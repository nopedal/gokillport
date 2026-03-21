# gokillport

Find and kill processes listening on given ports. One command, no noise.

## Install

```bash
go install github.com/nikolasopedal/gokillport/cmd/gokillport@latest
```

### Homebrew (coming soon)

```bash
brew tap nikolasopedal/tap
brew install gokillport
```

### From source

```bash
git clone https://github.com/nikolasopedal/gokillport.git
cd gokillport
make install
```

## Usage

Kill a process on a port:

```bash
gokillport 3000
```

Kill multiple ports at once:

```bash
gokillport 3000 8080 5432
```

Force kill (SIGKILL):

```bash
gokillport -f 3000
```

Dry run — see what would be killed:

```bash
gokillport -n 3000
```

List all listening processes:

```bash
gokillport -l
```

JSON output for scripting:

```bash
gokillport --json 3000
```

Quiet mode (errors only):

```bash
gokillport -q 3000
```

Custom timeout before escalation to SIGKILL:

```bash
gokillport -t 10 3000
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

## Build

```bash
make build          # current platform
make cross          # all platforms → dist/
make install        # install to $GOPATH/bin
make test           # run tests
```

## License

MIT
