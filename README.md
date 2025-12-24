# portlens

A lightweight local network traffic sniffer for Linux. Captures TCP/UDP traffic with process identification, connection tracking, and performance statistics.

## Features

- **Packet capture** using AF_PACKET sockets (no libpcap dependency)
- **Manual protocol parsing** - Ethernet, IPv4, TCP, UDP headers
- **Process identification** - maps connections to PIDs via /proc
- **Connection state tracking** - TCP state machine (SYN, ESTABLISHED, FIN, etc.)
- **JSON output** - structured, scriptable output format
- **YAML configuration** - persistent settings via config file
- **Performance statistics** - packets/sec, bytes/sec metrics
- **Graceful shutdown** - summary stats on Ctrl+C

## Requirements

- Linux (tested on Fedora 43, x86_64)
- Go 1.21+
- Root privileges (sudo)

## Installation

```bash
git clone https://github.com/hwang-fu/portlens.git
cd portlens
make build
```

## Quick Start

```bash
# Capture all traffic on loopback
sudo ./portlens -i lo

# Filter by protocol
sudo ./portlens -i eth0 --protocol tcp

# Filter by port
sudo ./portlens -i eth0 -p 8080

# Filter by process name
sudo ./portlens -i eth0 --process firefox

# Enable connection state tracking
sudo ./portlens -i lo --stateful

# Save output to file
sudo ./portlens -i lo -o capture.json

# Enable debug logging and performance stats
sudo ./portlens -i lo --debug --stats --graceful
```

## CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-i, --interface` | Network interface to capture on | (required) |
| `--protocol` | Protocol filter: tcp, udp, all | all |
| `-p, --port` | Filter by port number | 0 (all) |
| `--ip` | Filter by IP address | (all) |
| `--direction` | Filter: in, out, all | all |
| `--process` | Filter by process name | (all) |
| `--pid` | Filter by process ID | (all) |
| `--stateful` | Enable connection state tracking | false |
| `-v, --verbosity` | Output level: 0-3 | 2 |
| `-o, --output` | Write JSON to file | stdout |
| `-c, --config` | Config file path | ~/.config/portlens/config.yaml |
| `--debug` | Enable debug logging | false |
| `--log-file` | Write logs to file | stderr |
| `--stats` | Show performance statistics | false |
| `--graceful` | Clean shutdown with summary | false |
| `--version` | Show version | |

## Verbosity Levels

| Level | Output |
|-------|--------|
| 0 | Connection events only (requires --stateful) |
| 1 | Same as 0 |
| 2 | Individual packets (default) |
| 3 | Packets with payload preview |

## Configuration File

Create `~/.config/portlens/config.yaml`:

```yaml
interface: eth0
protocol: tcp
verbosity: 2
debug: false
stateful: true
```

CLI flags override config file values.

## Output Format

### Packet Record

```json
{
  "timestamp": "2025-12-24T10:30:45.123Z",
  "protocol": "TCP",
  "src_ip": "192.168.1.100",
  "src_port": 54321,
  "dst_ip": "93.184.216.34",
  "dst_port": 80,
  "direction": "out",
  "pid": 1234,
  "process": "curl",
  "tcp": {
    "seq": 123456,
    "ack": 789012,
    "flags": "SYN,ACK"
  }
}
```

### Connection Event (--stateful)

```json
{
  "event_type": "opened",
  "timestamp": "2025-12-24T10:30:45.123Z",
  "connection": {
    "src_ip": "192.168.1.100",
    "src_port": 54321,
    "dst_ip": "93.184.216.34",
    "dst_port": 80,
    "protocol": "TCP",
    "state": "ESTABLISHED",
    "packets_sent": 5,
    "packets_recv": 3,
    "bytes_sent": 1024,
    "bytes_recv": 2048
  }
}
```

### Stats (--stats)

```json
{
  "type": "stats",
  "timestamp": "2025-12-24T10:30:50.000Z",
  "elapsed_seconds": 5.0,
  "packets_captured": 100,
  "bytes_processed": 65000,
  "packets_per_sec": 20.0,
  "bytes_per_sec": 13000.0
}
```

## Testing

### Manual Testing

**Terminal 1:** Start portlens

```bash
sudo ./portlens -i lo --stats --graceful -v 3
```

**Terminal 2:** Generate traffic

```bash
# UDP traffic
echo "hello" | nc -u 127.0.0.1 12345

# TCP traffic (if you have a local server)
curl http://localhost:8080

# Multiple packets
for i in {1..10}; do echo "test $i" | nc -u 127.0.0.1 12345; sleep 1; done
```

**Terminal 1:** Press Ctrl+C to see shutdown summary.

### Unit Tests

```bash
make test
```

## Project Structure

```
portlens/
├── cmd/portlens/          # Entry point
│   └── main.go
├── internal/
│   ├── capture/           # AF_PACKET socket handling
│   ├── config/            # YAML config parsing
│   ├── output/            # JSON output structs
│   ├── parser/            # Protocol parsing (Ethernet, IPv4, TCP, UDP)
│   ├── procfs/            # Process identification via /proc
│   ├── stats/             # Performance statistics
│   └── tracker/           # Connection state tracking
├── Makefile
├── go.mod
└── README.md
```

## License

MIT
