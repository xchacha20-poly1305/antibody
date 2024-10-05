package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/XinRoom/iprange"
	E "github.com/sagernet/sing/common/exceptions"
	N "github.com/sagernet/sing/common/network"
	"github.com/xchacha20-poly1305/anchor"
	"github.com/xchacha20-poly1305/antibody"
	"golang.org/x/sync/errgroup"
)

func scanAnchor(ctx context.Context, iters []*iprange.Iter) context.Context {
	query, _ := anchor.Query{
		Version:    anchor.Version,
		DeviceName: "unknown",
	}.MarshalBinary()

	listenConfig := &net.ListenConfig{}
	pool := sync.Pool{
		New: func() any {
			conn, err := listenConfig.ListenPacket(ctx, N.NetworkUDP+"4", "")
			if err != nil {
				panic(err)
			}
			return conn
		},
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(threads)
	go func() {
		for _, iter := range iters {
			for i := uint64(0); i < iter.TotalNum(); i++ {
				ip := net.IP(bytes.Clone(iter.GetIpByIndex(i)))
				group.Go(func() error {
					conn := pool.Get().(net.PacketConn)
					defer pool.Put(conn)
					deadline := time.Now().Add(timeout)
					_ = conn.SetDeadline(deadline)
					probeCtx, cancelProbe := context.WithDeadline(ctx, deadline)
					response, err := antibody.ProbeAnchor(probeCtx, ip, conn, query)
					cancelProbe()
					if err != nil {
						if debugMode && !E.IsClosedOrCanceled(err) && !E.IsTimeout(err) {
							fmt.Fprintf(os.Stderr, "[%s] is not anchor: %v\n", ip, err)
						}
						return nil
					}
					fmt.Printf("[%s] is anchor: %+v\n", ip, response)
					return nil
				})
			}
		}
		_ = group.Wait()
	}()
	return groupCtx
}
