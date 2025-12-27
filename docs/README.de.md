# portlens

[English](../README.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [繁體中文](README.zh.md) | [日本語](README.jp.md)

> **Hinweis:** Diese README wurde ursprünglich auf Englisch verfasst. Bei Unklarheiten konsultieren Sie bitte die [englische Version](../README.md).

Ein leichtgewichtiger lokaler Netzwerk-Traffic-Sniffer für Linux. Erfasst TCP/UDP-Verkehr mit Prozessidentifikation, Verbindungsverfolgung und Leistungsstatistiken.

## Funktionen

- **Paketerfassung** mittels AF_PACKET-Sockets (keine libpcap-Abhängigkeit)
- **Manuelle Protokollanalyse** - Ethernet, IPv4, TCP, UDP Header
- **Prozessidentifikation** - ordnet Verbindungen PIDs über /proc zu
- **Verbindungszustandsverfolgung** - TCP-Zustandsmaschine (SYN, ESTABLISHED, FIN, etc.)
- **JSON-Ausgabe** - strukturiertes, skriptfähiges Ausgabeformat
- **YAML-Konfiguration** - dauerhafte Einstellungen über Konfigurationsdatei
- **Leistungsstatistiken** - Pakete/Sek., Bytes/Sek. Metriken
- **Sauberes Beenden** - Zusammenfassungsstatistiken bei Ctrl+C

## Anforderungen

- Linux (getestet auf Fedora 43, x86_64)
- Go 1.21+
- Root-Rechte (sudo)

## Installation

```bash
git clone https://github.com/hwang-fu/portlens.git
cd portlens
make build
```

## Schnellstart

```bash
# Gesamten Verkehr auf Loopback erfassen
sudo ./portlens -i lo

# Nach Protokoll filtern
sudo ./portlens -i eth0 --protocol tcp

# Nach Port filtern
sudo ./portlens -i eth0 -p 8080

# Nach Prozessname filtern
sudo ./portlens -i eth0 --process firefox

# Verbindungszustandsverfolgung aktivieren
sudo ./portlens -i lo --stateful

# Ausgabe in Datei speichern
sudo ./portlens -i lo -o capture.json

# Debug-Protokollierung und Leistungsstatistiken aktivieren
sudo ./portlens -i lo --debug --stats --graceful
```

## CLI-Parameter

| Parameter | Beschreibung | Standard |
|-----------|--------------|----------|
| `-i, --interface` | Netzwerk-Schnittstelle zur Erfassung | (erforderlich) |
| `--protocol` | Protokollfilter: tcp, udp, all | all |
| `-p, --port` | Nach Portnummer filtern | 0 (alle) |
| `--ip` | Nach IP-Adresse filtern | (alle) |
| `--direction` | Filter: in, out, all | all |
| `--process` | Nach Prozessname filtern | (alle) |
| `--pid` | Nach Prozess-ID filtern | (alle) |
| `--stateful` | Verbindungszustandsverfolgung aktivieren | false |
| `-v, --verbosity` | Ausgabestufe: 0-3 | 2 |
| `-o, --output` | JSON in Datei schreiben | stdout |
| `-c, --config` | Konfigurationsdateipfad | ~/.config/portlens/config.yaml |
| `--debug` | Debug-Protokollierung aktivieren | false |
| `--log-file` | Protokolle in Datei schreiben | stderr |
| `--stats` | Leistungsstatistiken anzeigen | false |
| `--graceful` | Sauberes Beenden mit Zusammenfassung | false |
| `--version` | Version anzeigen | |

## Ausführlichkeitsstufen

| Stufe | Ausgabe |
|-------|---------|
| 0 | Nur Verbindungsereignisse (erfordert --stateful) |
| 1 | Wie 0 |
| 2 | Einzelne Pakete (Standard) |
| 3 | Pakete mit Payload-Vorschau |

## Konfigurationsdatei

Erstellen Sie `~/.config/portlens/config.yaml`:

```yaml
interface: eth0
protocol: tcp
verbosity: 2
debug: false
stateful: true
```

CLI-Parameter überschreiben Konfigurationsdateiwerte.

## Ausgabeformat

### Paketdatensatz

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

### Verbindungsereignis (--stateful)

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

### Statistiken (--stats)

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

## Testen

### Manuelles Testen

**Terminal 1:** portlens starten

```bash
sudo ./portlens -i lo --stats --graceful -v 3
```

**Terminal 2:** Verkehr erzeugen

```bash
# UDP-Verkehr
echo "hello" | nc -u 127.0.0.1 12345

# TCP-Verkehr (wenn lokaler Server vorhanden)
curl http://localhost:8080

# Mehrere Pakete
for i in {1..10}; do echo "test $i" | nc -u 127.0.0.1 12345; sleep 1; done
```

**Terminal 1:** Ctrl+C drücken für Beendigungszusammenfassung.

### Unit-Tests

```bash
make test
```

## Projektstruktur

```
portlens/
├── cmd/portlens/          # Einstiegspunkt
│   └── main.go
├── internal/
│   ├── capture/           # AF_PACKET-Socket-Behandlung
│   ├── config/            # YAML-Konfigurationsanalyse
│   ├── output/            # JSON-Ausgabestrukturen
│   ├── parser/            # Protokollanalyse (Ethernet, IPv4, TCP, UDP)
│   ├── procfs/            # Prozessidentifikation über /proc
│   ├── stats/             # Leistungsstatistiken
│   └── tracker/           # Verbindungszustandsverfolgung
├── Makefile
├── go.mod
└── README.md
```

## Lizenz

MIT
