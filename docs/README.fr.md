# portlens

[English](../README.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [繁體中文](README.zh.md) | [日本語](README.jp.md)

> **Note :** Ce README a été rédigé à l'origine en anglais. En cas de doute, veuillez consulter la [version anglaise](../README.md).

Un analyseur de trafic réseau local léger pour Linux. Capture le trafic TCP/UDP avec identification des processus, suivi des connexions et statistiques de performance.

## Fonctionnalités

- **Capture de paquets** utilisant les sockets AF_PACKET (sans dépendance libpcap)
- **Analyse manuelle des protocoles** - En-têtes Ethernet, IPv4, TCP, UDP
- **Identification des processus** - associe les connexions aux PIDs via /proc
- **Suivi de l'état des connexions** - Machine d'état TCP (SYN, ESTABLISHED, FIN, etc.)
- **Sortie JSON** - format de sortie structuré et scriptable
- **Configuration YAML** - paramètres persistants via fichier de configuration
- **Statistiques de performance** - métriques paquets/sec, octets/sec
- **Arrêt propre** - statistiques récapitulatives sur Ctrl+C

## Prérequis

- Linux (testé sur Fedora 43, x86_64)
- Go 1.21+
- Privilèges root (sudo)

## Installation

```bash
git clone https://github.com/hwang-fu/portlens.git
cd portlens
make build
```

## Démarrage rapide

```bash
# Capturer tout le trafic sur loopback
sudo ./portlens -i lo

# Filtrer par protocole
sudo ./portlens -i eth0 --protocol tcp

# Filtrer par port
sudo ./portlens -i eth0 -p 8080

# Filtrer par nom de processus
sudo ./portlens -i eth0 --process firefox

# Activer le suivi de l'état des connexions
sudo ./portlens -i lo --stateful

# Enregistrer la sortie dans un fichier
sudo ./portlens -i lo -o capture.json

# Activer la journalisation de débogage et les statistiques de performance
sudo ./portlens -i lo --debug --stats --graceful
```

## Options CLI

| Option | Description | Défaut |
|--------|-------------|--------|
| `-i, --interface` | Interface réseau à capturer | (requis) |
| `--protocol` | Filtre de protocole : tcp, udp, all | all |
| `-p, --port` | Filtrer par numéro de port | 0 (tous) |
| `--ip` | Filtrer par adresse IP | (tous) |
| `--direction` | Filtre : in, out, all | all |
| `--process` | Filtrer par nom de processus | (tous) |
| `--pid` | Filtrer par ID de processus | (tous) |
| `--stateful` | Activer le suivi de l'état des connexions | false |
| `-v, --verbosity` | Niveau de sortie : 0-3 | 2 |
| `-o, --output` | Écrire JSON dans un fichier | stdout |
| `-c, --config` | Chemin du fichier de configuration | ~/.config/portlens/config.yaml |
| `--debug` | Activer la journalisation de débogage | false |
| `--log-file` | Écrire les journaux dans un fichier | stderr |
| `--stats` | Afficher les statistiques de performance | false |
| `--graceful` | Arrêt propre avec résumé | false |
| `--version` | Afficher la version | |

## Niveaux de verbosité

| Niveau | Sortie |
|--------|--------|
| 0 | Événements de connexion uniquement (nécessite --stateful) |
| 1 | Identique à 0 |
| 2 | Paquets individuels (défaut) |
| 3 | Paquets avec aperçu du payload |

## Fichier de configuration

Créez `~/.config/portlens/config.yaml` :

```yaml
interface: eth0
protocol: tcp
verbosity: 2
debug: false
stateful: true
```

Les options CLI remplacent les valeurs du fichier de configuration.

## Format de sortie

### Enregistrement de paquet

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

### Événement de connexion (--stateful)

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

### Statistiques (--stats)

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

## Tests

### Tests manuels

**Terminal 1 :** Démarrer portlens

```bash
sudo ./portlens -i lo --stats --graceful -v 3
```

**Terminal 2 :** Générer du trafic

```bash
# Trafic UDP
echo "hello" | nc -u 127.0.0.1 12345

# Trafic TCP (si vous avez un serveur local)
curl http://localhost:8080

# Plusieurs paquets
for i in {1..10}; do echo "test $i" | nc -u 127.0.0.1 12345; sleep 1; done
```

**Terminal 1 :** Appuyez sur Ctrl+C pour voir le résumé d'arrêt.

### Tests unitaires

```bash
make test
```

## Structure du projet

```
portlens/
├── cmd/portlens/          # Point d'entrée
│   └── main.go
├── internal/
│   ├── capture/           # Gestion des sockets AF_PACKET
│   ├── config/            # Analyse de configuration YAML
│   ├── output/            # Structures de sortie JSON
│   ├── parser/            # Analyse des protocoles (Ethernet, IPv4, TCP, UDP)
│   ├── procfs/            # Identification des processus via /proc
│   ├── stats/             # Statistiques de performance
│   └── tracker/           # Suivi de l'état des connexions
├── Makefile
├── go.mod
└── README.md
```

## Licence

MIT
