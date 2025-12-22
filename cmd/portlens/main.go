package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/hwang-fu/portlens/internal/capture"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: portlens <interface>")
		fmt.Println("example: sudo ./portlens eth0")
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
		fmt.Printf("\n=== Packet: %d bytes ===\n", n)
		fmt.Println(hex.Dump(buf[:n]))
	}
}
