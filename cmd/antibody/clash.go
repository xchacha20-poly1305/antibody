package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"

	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/xchacha20-poly1305/antibody"
	"golang.org/x/sync/errgroup"
)

func scanClash(ctx context.Context, targets []WithPorts) {
	pool := sync.Pool{
		New: func() any {
			return &http.Client{}
		},
	}

	group, _ := errgroup.WithContext(ctx)
	group.SetLimit(threads)
	for _, target := range targets {
		ipString := target.ip.String()
		for _, port := range target.ports {
			port := port
			group.Go(func() error {
				client := pool.Get().(*http.Client)
				defer pool.Put(client)
				probeCtx, cancelProbe := context.WithTimeout(ctx, timeout)
				host := net.JoinHostPort(ipString, F.ToString(port))
				info, err := antibody.ProbeClash(probeCtx, client, &url.URL{
					Scheme: "http",
					Host:   host,
				})
				cancelProbe()
				if err != nil {
					if debugMode && !E.IsClosedOrCanceled(err) {
						fmt.Fprintf(os.Stderr, "[%s] is not clash API: %v\n", host, err)
					}
					return nil
				}
				fmt.Printf("[%s] is clash: %+v\n", host, info)
				return nil
			})
		}
	}
	_ = group.Wait()
}
