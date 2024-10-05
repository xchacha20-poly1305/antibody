package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/XinRoom/iprange"
)

const (
	ModeAnchor = "anchor"
	ModeClash  = "clash"
)

var (
	mode           string
	rawIP          string
	tcpScanTimeout time.Duration
	timeout        time.Duration
	threads        int
	debugMode      bool
)

func main() {
	flag.StringVar(&mode, "m", ModeAnchor, "Scan mode. Support: anchor, clash")
	flag.StringVar(&rawIP, "i", "192.168.0.0/16", "The CIDR you want to scan. Split by \",\"")
	flag.DurationVar(&tcpScanTimeout, "tt", 500*time.Second, "TCP scan timeout")
	flag.DurationVar(&timeout, "t", 3*time.Second, "Timeout")
	flag.IntVar(&threads, "x", runtime.NumCPU()*8, "Threads.")
	flag.BoolVar(&debugMode, "d", false, "Debug mode")
	flag.Parse()

	ranges := strings.Split(rawIP, ",")
	iters := make([]*iprange.Iter, 0, len(ranges))
	for _, rawRange := range ranges {
		iter, _, err := iprange.NewIter(rawRange)
		if err != nil {
			fatal("Parse %s: %v\n", rawRange, err)
		}
		iters = append(iters, iter)
	}

	switch mode {
	case ModeAnchor:
		scanAnchor(context.Background(), iters)
	case ModeClash:
		var targets []WithPorts
		for _, iter := range iters {
			for i := uint64(0); i < iter.TotalNum(); i++ {
				ip := net.IP(bytes.Clone(iter.GetIpByIndex(i)))
				ports := scanTcp(context.Background(), ip.String())
				if len(ports) == 0 {
					continue
				}
				targets = append(targets, WithPorts{
					ip:    ip,
					ports: ports,
				})
			}
		}
		scanClash(context.Background(), targets)
	default:
		fatal("Unknown mode: %s\v", mode)
	}
}

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
