package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hwang-fu/portlens/internal/capture"
	"github.com/hwang-fu/portlens/internal/output"
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

	fmt.Fprintf(os.Stderr, "capturing on %s...\n", interfaceName)

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

		ipv4Packet, err := parser.ParseIPv4(frame.Payload)
		if err != nil {
			log.Printf("parse ipv4 error: %v", err)
			continue
		}

		switch ipv4Packet.Protocol {
		case parser.ProtocolTCP:
			tcpSegment, err := parser.ParseTCP(ipv4Packet.Payload)
			if err != nil {
				log.Printf("parse TCP error: %v", err)
				continue
			}
			record := output.PacketRecord{
				Timestamp: output.Now(),
				Protocol:  "TCP",
				SrcIP:     ipv4Packet.SrcIP.String(),
				SrcPort:   tcpSegment.SrcPort,
				DstIP:     ipv4Packet.DstIP.String(),
				DstPort:   tcpSegment.DstPort,
				TCP: &output.TCPInfo{
					Seq:   tcpSegment.SeqNum,
					Ack:   tcpSegment.AckNum,
					Flags: parser.FormatFlags(tcpSegment.Flags),
				},
			}
			json.NewEncoder(os.Stdout).Encode(record)

		case parser.ProtocolUDP:
			udpDatagram, err := parser.ParseUDP(ipv4Packet.Payload)
			if err != nil {
				log.Printf("parse UDP error: %v", err)
				continue
			}
			record := output.PacketRecord{
				Timestamp: output.Now(),
				Protocol:  "UDP",
				SrcIP:     ipv4Packet.SrcIP.String(),
				SrcPort:   udpDatagram.SrcPort,
				DstIP:     ipv4Packet.DstIP.String(),
				DstPort:   udpDatagram.DstPort,
				UDP: &output.UDPInfo{
					Length: udpDatagram.Length,
				},
			}
			json.NewEncoder(os.Stdout).Encode(record)
		default:
			// Skip non-TCP/UDP packets
			continue
		}
	}
}
