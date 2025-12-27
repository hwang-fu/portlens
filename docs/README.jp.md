# portlens

[English](../README.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [繁體中文](README.zh.md) | [日本語](README.jp.md)

> **注意：** このREADMEは元々英語で書かれています。内容が不明な場合は、[英語版](../README.md)をご参照ください。

Linux用の軽量なローカルネットワークトラフィックスニッファー。プロセス識別、接続追跡、パフォーマンス統計機能を備えたTCP/UDPトラフィックをキャプチャします。

## 機能

- **パケットキャプチャ** AF_PACKETソケット使用（libpcap依存なし）
- **手動プロトコル解析** - Ethernet、IPv4、TCP、UDPヘッダー
- **プロセス識別** - /proc経由で接続をPIDにマッピング
- **接続状態追跡** - TCPステートマシン（SYN、ESTABLISHED、FINなど）
- **JSON出力** - 構造化されたスクリプト可能な出力形式
- **YAML設定** - 設定ファイルによる永続的な設定
- **パフォーマンス統計** - パケット/秒、バイト/秒メトリクス
- **グレースフルシャットダウン** - Ctrl+Cで概要統計を表示

## 要件

- Linux（Fedora 43, x86_64でテスト済み）
- Go 1.21+
- root権限（sudo）

## インストール

```bash
git clone https://github.com/hwang-fu/portlens.git
cd portlens
make build
```

## クイックスタート

```bash
# ループバックのすべてのトラフィックをキャプチャ
sudo ./portlens -i lo

# プロトコルでフィルタリング
sudo ./portlens -i eth0 --protocol tcp

# ポートでフィルタリング
sudo ./portlens -i eth0 -p 8080

# プロセス名でフィルタリング
sudo ./portlens -i eth0 --process firefox

# 接続状態追跡を有効化
sudo ./portlens -i lo --stateful

# 出力をファイルに保存
sudo ./portlens -i lo -o capture.json

# デバッグログとパフォーマンス統計を有効化
sudo ./portlens -i lo --debug --stats --graceful
```

## CLIフラグ

| フラグ | 説明 | デフォルト |
|--------|------|------------|
| `-i, --interface` | キャプチャするネットワークインターフェース | （必須） |
| `--protocol` | プロトコルフィルタ：tcp、udp、all | all |
| `-p, --port` | ポート番号でフィルタリング | 0（すべて） |
| `--ip` | IPアドレスでフィルタリング | （すべて） |
| `--direction` | フィルタ：in、out、all | all |
| `--process` | プロセス名でフィルタリング | （すべて） |
| `--pid` | プロセスIDでフィルタリング | （すべて） |
| `--stateful` | 接続状態追跡を有効化 | false |
| `-v, --verbosity` | 出力レベル：0-3 | 2 |
| `-o, --output` | JSONをファイルに書き込み | stdout |
| `-c, --config` | 設定ファイルパス | ~/.config/portlens/config.yaml |
| `--debug` | デバッグログを有効化 | false |
| `--log-file` | ログをファイルに書き込み | stderr |
| `--stats` | パフォーマンス統計を表示 | false |
| `--graceful` | 概要付きでクリーンシャットダウン | false |
| `--version` | バージョンを表示 | |

## 詳細レベル

| レベル | 出力 |
|--------|------|
| 0 | 接続イベントのみ（--stateful必須） |
| 1 | 0と同じ |
| 2 | 個別パケット（デフォルト） |
| 3 | パケット + ペイロードプレビュー |

## 設定ファイル

`~/.config/portlens/config.yaml`を作成：

```yaml
interface: eth0
protocol: tcp
verbosity: 2
debug: false
stateful: true
```

CLIフラグは設定ファイルの値を上書きします。

## 出力形式

### パケットレコード

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

### 接続イベント（--stateful）

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

### 統計（--stats）

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

## テスト

### 手動テスト

**ターミナル1：** portlensを起動

```bash
sudo ./portlens -i lo --stats --graceful -v 3
```

**ターミナル2：** トラフィックを生成

```bash
# UDPトラフィック
echo "hello" | nc -u 127.0.0.1 12345

# TCPトラフィック（ローカルサーバーがある場合）
curl http://localhost:8080

# 複数パケット
for i in {1..10}; do echo "test $i" | nc -u 127.0.0.1 12345; sleep 1; done
```

**ターミナル1：** Ctrl+Cを押してシャットダウン概要を表示。

### ユニットテスト

```bash
make test
```

## プロジェクト構造

```
portlens/
├── cmd/portlens/          # エントリーポイント
│   └── main.go
├── internal/
│   ├── capture/           # AF_PACKETソケット処理
│   ├── config/            # YAML設定解析
│   ├── output/            # JSON出力構造体
│   ├── parser/            # プロトコル解析（Ethernet、IPv4、TCP、UDP）
│   ├── procfs/            # /proc経由のプロセス識別
│   ├── stats/             # パフォーマンス統計
│   └── tracker/           # 接続状態追跡
├── Makefile
├── go.mod
└── README.md
```

## ライセンス

MIT
