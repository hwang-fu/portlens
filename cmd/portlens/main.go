package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hwang-fu/portlens/internal/capture"
	"github.com/hwang-fu/portlens/internal/parser"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: portlens <interface>")
		// Use `ip link show` to list available interfaces.
		// To test with loopback: run `sudo ./portlens lo` then `ping 127.0.0.1` in another terminal.
		fmt.Println("example: sudo ./portlens lo")
		os.Exit(1)
	}

	interfaceName := os.Args[1]
	sock, err := capture.NewSocket()
	if err != nil {
		log.Fatalf("create socket: %v", err)
	}
	defer sock.Close()

	if err := sock.Bind(interfaceName); err != nil {
		log.Fatalf("bind: %v", err)
	}

	fmt.Printf("capturing on %s...\n", interfaceName)

	buf := make([]byte, 65535)
	for {
		n, err := sock.ReadPacket(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		frame, err := parser.ParseEthernet(buf[:n])
		if err != nil {
			log.Printf("parse error: %v", err)
			continue
		}
		// Skip non-IPv4 packets by far
		if frame.EtherType != parser.EtherTypeIPv4 {
			continue
		}

		fmt.Printf("\n=== Packet: %d bytes ===\n", n)
		fmt.Printf("Src MAC: %s\n", frame.SrcMAC)
		fmt.Printf("Dst MAC: %s\n", frame.DestMAC)
		fmt.Printf("Payload: %d bytes\n", len(frame.Payload))
	}
}
