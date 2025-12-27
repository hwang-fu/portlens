# portlens

[English](../README.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [繁體中文](README.zh.md) | [日本語](README.jp.md)

> **注意：** 本 README 原文為英文撰寫。如有任何內容不清楚之處，請參閱[英文版本](../README.md)。

一個輕量級的 Linux 本地網路流量嗅探器。可捕獲 TCP/UDP 流量，並提供進程識別、連線追蹤和效能統計功能。

## 功能特色

- **封包捕獲** 使用 AF_PACKET socket（無需 libpcap 依賴）
- **手動協定解析** - Ethernet、IPv4、TCP、UDP 標頭
- **進程識別** - 透過 /proc 將連線對應到 PID
- **連線狀態追蹤** - TCP 狀態機（SYN、ESTABLISHED、FIN 等）
- **JSON 輸出** - 結構化、可腳本化的輸出格式
- **YAML 設定** - 透過設定檔儲存持久設定
- **效能統計** - 封包/秒、位元組/秒指標
- **優雅關閉** - Ctrl+C 時顯示摘要統計

## 系統需求

- Linux（已在 Fedora 43, x86_64 上測試）
- Go 1.21+
- Root 權限（sudo）

## 安裝

```bash
git clone https://github.com/hwang-fu/portlens.git
cd portlens
make build
```

## 快速開始

```bash
# 捕獲 loopback 上的所有流量
sudo ./portlens -i lo

# 按協定過濾
sudo ./portlens -i eth0 --protocol tcp

# 按埠號過濾
sudo ./portlens -i eth0 -p 8080

# 按進程名稱過濾
sudo ./portlens -i eth0 --process firefox

# 啟用連線狀態追蹤
sudo ./portlens -i lo --stateful

# 將輸出儲存到檔案
sudo ./portlens -i lo -o capture.json

# 啟用除錯日誌和效能統計
sudo ./portlens -i lo --debug --stats --graceful
```

## CLI 參數

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-i, --interface` | 要捕獲的網路介面 | （必填） |
| `--protocol` | 協定過濾器：tcp、udp、all | all |
| `-p, --port` | 按埠號過濾 | 0（全部） |
| `--ip` | 按 IP 位址過濾 | （全部） |
| `--direction` | 過濾：in、out、all | all |
| `--process` | 按進程名稱過濾 | （全部） |
| `--pid` | 按進程 ID 過濾 | （全部） |
| `--stateful` | 啟用連線狀態追蹤 | false |
| `-v, --verbosity` | 輸出級別：0-3 | 2 |
| `-o, --output` | 將 JSON 寫入檔案 | stdout |
| `-c, --config` | 設定檔路徑 | ~/.config/portlens/config.yaml |
| `--debug` | 啟用除錯日誌 | false |
| `--log-file` | 將日誌寫入檔案 | stderr |
| `--stats` | 顯示效能統計 | false |
| `--graceful` | 優雅關閉並顯示摘要 | false |
| `--version` | 顯示版本 | |

## 詳細程度級別

| 級別 | 輸出 |
|------|------|
| 0 | 僅連線事件（需要 --stateful） |
| 1 | 與 0 相同 |
| 2 | 個別封包（預設） |
| 3 | 封包及負載預覽 |

## 設定檔

建立 `~/.config/portlens/config.yaml`：

```yaml
interface: eth0
protocol: tcp
verbosity: 2
debug: false
stateful: true
```

CLI 參數會覆蓋設定檔中的值。

## 輸出格式

### 封包記錄

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

### 連線事件（--stateful）

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

## 測試

### 手動測試

**終端機 1：** 啟動 portlens

```bash
sudo ./portlens -i lo --stats --graceful -v 3
```

**終端機 2：** 產生流量

```bash
# UDP 流量
echo "hello" | nc -u 127.0.0.1 12345

# TCP 流量（如果你有本地伺服器）
curl http://localhost:8080

# 多個封包
for i in {1..10}; do echo "test $i" | nc -u 127.0.0.1 12345; sleep 1; done
```

**終端機 1：** 按 Ctrl+C 查看關閉摘要。

### 單元測試

```bash
make test
```

## 專案結構

```
portlens/
├── cmd/portlens/          # 進入點
│   └── main.go
├── internal/
│   ├── capture/           # AF_PACKET socket 處理
│   ├── config/            # YAML 設定解析
│   ├── output/            # JSON 輸出結構
│   ├── parser/            # 協定解析（Ethernet、IPv4、TCP、UDP）
│   ├── procfs/            # 透過 /proc 進行進程識別
│   ├── stats/             # 效能統計
│   └── tracker/           # 連線狀態追蹤
├── Makefile
├── go.mod
└── README.md
```

## 授權條款

MIT
