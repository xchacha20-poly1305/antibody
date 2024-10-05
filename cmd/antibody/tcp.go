package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync"

	F "github.com/sagernet/sing/common/format"
	N "github.com/sagernet/sing/common/network"
	"golang.org/x/sync/errgroup"
)

func scanTcp(ctx context.Context, ip string) (ports []uint16) {
	group, _ := errgroup.WithContext(ctx)
	pool := sync.Pool{
		New: func() any {
			return &net.Dialer{
				Timeout: tcpScanTimeout,
			}
		},
	}
	var mu sync.Mutex
	for i := uint16(80); i < uint16(math.MaxUint16); i++ {
		i := i
		group.Go(func() error {
			dialer := pool.Get().(*net.Dialer)
			defer pool.Put(dialer)
			host := net.JoinHostPort(ip, F.ToString(i))
			conn, err := dialer.DialContext(ctx, N.NetworkTCP, host)
			if err != nil {
				if debugMode {
					// fmt.Fprintf(os.Stderr, "[%s] is not opened: %v\n", host, err)
				}
				return nil
			}
			if debugMode {
				fmt.Printf("[%s] is providing TCP\n", host)
			}
			defer conn.Close()
			mu.Lock()
			defer mu.Unlock()
			ports = append(ports, i)
			return nil
		})
	}
	_ = group.Wait()
	return
}

type WithPorts struct {
	ip    net.IP
	ports []uint16
}
