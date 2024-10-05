package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/XinRoom/iprange"
)

const (
	ModeAnchor = "anchor"
)

var (
	mode      string
	rawIP     string
	timeout   time.Duration
	threads   int
	debugMode bool
)

func main() {
	flag.StringVar(&mode, "m", ModeAnchor, "Scan mode. Support: anchor")
	flag.StringVar(&rawIP, "i", "192.168.0.0/16", "The CIDR you want to scan. Split by \",\"")
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	var groupCtx context.Context
	switch mode {
	case ModeAnchor:
		groupCtx = scanAnchor(ctx, iters)
	default:
		fatal("Unknown mode: %s\v", mode)
	}
	select {
	case <-ctx.Done():
		return
	case <-groupCtx.Done():
		cancel()
	}
}

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
