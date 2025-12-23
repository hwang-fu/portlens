package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hwang-fu/portlens/internal/capture"
	"github.com/hwang-fu/portlens/internal/output"
	"github.com/hwang-fu/portlens/internal/parser"
)

var version = "dev"

func main() {
	// Define flags
	var (
		interfaceName = flag.String("interface", "", "network interface to capture on")
		protocol      = flag.String("protocol", "all", "protocol to capture: tcp, udp, or all")
		showVersion   = flag.Bool("version", false, "show version and exit")
		port          = flag.Int("port", 0, "filter by port number (0 = all ports)")
		ip            = flag.String("ip", "", "filter by IP address (empty = all IPs)")
	)

	// Short aliases
	flag.StringVar(interfaceName, "i", "", "network interface (shorthand)")
	flag.IntVar(port, "p", 0, "filter by port (shorthand)")

	flag.Parse()

	if *showVersion {
		fmt.Println("portlens", version)
		os.Exit(0)
	}

	if *interfaceName == "" {
		fmt.Fprintln(os.Stderr, "error: --interface (-i) is required")
		fmt.Fprintln(os.Stderr, "usage: portlens -i <interface> [--protocol tcp|udp|all]")
		fmt.Fprintln(os.Stderr, "example: sudo portlens -i lo")
		os.Exit(1)
	}

	sock, err := capture.NewSocket()
	if err != nil {
		log.Fatalf("create socket: %v", err)
	}
	defer sock.Close()

	if err := sock.Bind(*interfaceName); err != nil {
		log.Fatalf("bind: %v", err)
	}

	fmt.Fprintf(os.Stderr, "capturing on %s...\n", *interfaceName)

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

		// IP filter: check if src or dst IP matches the filtering ip (if provided)
		if *ip != "" && ipv4Packet.SrcIP.String() != *ip && ipv4Packet.DstIP.String() != *ip {
			continue
		}

		switch ipv4Packet.Protocol {
		case parser.ProtocolTCP:
			if *protocol == "udp" {
				continue // Skip TCP when filtering for UDP only
			}

			tcpSegment, err := parser.ParseTCP(ipv4Packet.Payload)
			if err != nil {
				log.Printf("parse TCP error: %v", err)
				continue
			}

			// Port filter: check if src or dst port matches
			if *port != 0 && tcpSegment.SrcPort != uint16(*port) && tcpSegment.DstPort != uint16(*port) {
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
			if *protocol == "tcp" {
				continue // Skip UDP when filtering for TCP only
			}

			udpDatagram, err := parser.ParseUDP(ipv4Packet.Payload)
			if err != nil {
				log.Printf("parse UDP error: %v", err)
				continue
			}

			// Port filter: check if src or dst port matches
			if *port != 0 && udpDatagram.SrcPort != uint16(*port) && udpDatagram.DstPort != uint16(*port) {
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
