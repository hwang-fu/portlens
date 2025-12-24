package main

import (
	"flag"
	"fmt"
	"os"

	yamlconfig "github.com/hwang-fu/portlens/internal/config"
)

// config holds all runtime configuration from flags.
type config struct {
	interfaceName string
	protocol      string
	port          int
	ip            string
	direction     string
	process       string
	pid           int
	stateful      bool
	verbosity     int    // 0=minimal, 1=normal, 2=detailed, 3=verbose
	outputFile    string // output file path (empty = stdout)
	debug         bool   // enable debug logging
	logFile       string // log file path (empty = stderr)
	configFile    string // config file path
	stats         bool   // show performance statistics
	graceful      bool   // enable graceful shutdown with summary
}

func parseFlags() {
	// Check for config file flag first (manual parse)
	configPath := yamlconfig.DefaultPath()
	for i, arg := range os.Args[1:] {
		if arg == "-c" || arg == "--config" {
			if i+1 < len(os.Args)-1 {
				configPath = os.Args[i+2]
			}
		}
	}

	// Load config file (if exists)
	fileCfg, err := yamlconfig.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// Set defaults from file config
	cfg.configFile = configPath
	cfg.interfaceName = fileCfg.Interface
	cfg.protocol = fileCfg.Protocol
	cfg.port = fileCfg.Port
	cfg.ip = fileCfg.IP
	cfg.direction = fileCfg.Direction
	cfg.process = fileCfg.Process
	cfg.pid = fileCfg.PID
	cfg.stateful = fileCfg.Stateful
	cfg.verbosity = fileCfg.Verbosity
	cfg.outputFile = fileCfg.Output
	cfg.debug = fileCfg.Debug
	cfg.logFile = fileCfg.LogFile
	cfg.stats = fileCfg.Stats
	cfg.graceful = fileCfg.Graceful

	// Default verbosity if not set
	if cfg.verbosity == 0 {
		cfg.verbosity = 2
	}
	// Default protocol if not set
	if cfg.protocol == "" {
		cfg.protocol = "all"
	}
	// Default direction if not set
	if cfg.direction == "" {
		cfg.direction = "all"
	}

	flag.StringVar(&cfg.interfaceName, "interface", cfg.interfaceName, "network interface to capture on")
	flag.StringVar(&cfg.interfaceName, "i", cfg.interfaceName, "network interface (shorthand)")
	flag.StringVar(&cfg.protocol, "protocol", cfg.protocol, "protocol to capture: tcp, udp, or all")
	flag.IntVar(&cfg.port, "port", cfg.port, "filter by port number (0 = all ports)")
	flag.IntVar(&cfg.port, "p", cfg.port, "filter by port (shorthand)")
	flag.StringVar(&cfg.ip, "ip", cfg.ip, "filter by IP address (empty = all IPs)")
	flag.StringVar(&cfg.direction, "direction", cfg.direction, "filter by direction: in, out, or all")
	flag.StringVar(&cfg.process, "process", cfg.process, "filter by process name")
	flag.IntVar(&cfg.pid, "pid", cfg.pid, "filter by process ID")
	flag.BoolVar(&cfg.stateful, "stateful", cfg.stateful, "enable connection state tracking")
	flag.IntVar(&cfg.verbosity, "verbosity", cfg.verbosity, "output verbosity: 0=minimal, 1=normal, 2=detailed, 3=verbose")
	flag.IntVar(&cfg.verbosity, "v", cfg.verbosity, "verbosity level (shorthand)")
	flag.StringVar(&cfg.outputFile, "output", cfg.outputFile, "write output to file (default: stdout)")
	flag.StringVar(&cfg.outputFile, "o", cfg.outputFile, "output file (shorthand)")
	flag.BoolVar(&cfg.debug, "debug", cfg.debug, "enable debug logging")
	flag.StringVar(&cfg.logFile, "log-file", cfg.logFile, "write logs to file (default: stderr)")
	flag.StringVar(&cfg.configFile, "config", cfg.configFile, "config file path")
	flag.StringVar(&cfg.configFile, "c", cfg.configFile, "config file (shorthand)")
	flag.BoolVar(&cfg.stats, "stats", cfg.stats, "show performance statistics")
	flag.BoolVar(&cfg.graceful, "graceful", cfg.graceful, "enable graceful shutdown with summary")

	showVersion := flag.Bool("version", false, "show version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Println("portlens", version)
		os.Exit(0)
	}

	if cfg.interfaceName == "" {
		fmt.Fprintln(os.Stderr, "error: --interface (-i) is required")
		fmt.Fprintln(os.Stderr, "usage: portlens -i <interface> [--protocol tcp|udp|all]")
		fmt.Fprintln(os.Stderr, "example: sudo portlens -i lo")
		os.Exit(1)
	}
}
