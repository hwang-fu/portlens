# portlens

A lightweight local network traffic sniffer for Linux. Captures TCP/UDP traffic and maps connections to processes.

## Requirements

- Linux (tested on Fedora 43)
- Go 1.25+
- Root privileges (sudo)

## Build

```bash
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
```

## Output

JSON to stdout (one object per line):

```json
{"timestamp":"2025-12-23T10:30:45.123Z","protocol":"TCP","src_ip":"127.0.0.1","src_port":8080,"dst_ip":"127.0.0.1","dst_port":54321,"direction":"out","pid":1234,"process":"curl","tcp":{"seq":123,"ack":456,"flags":"SYN,ACK"}}
```

## License

MIT
