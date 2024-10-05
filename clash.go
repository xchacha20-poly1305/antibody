package antibody

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
)

const (
	ClashTypeClassic = "Clash"
	ClashTypePremium = "Clash Premium"
	ClashTypeMihomo  = "Mihomo"
	ClashTypeBox     = "sing-box"
)

type ClashInfo struct {
	Type    string
	Version string
}

var requestPool = sync.Pool{
	New: func() any {
		return &http.Request{
			Method: http.MethodGet,
		}
	},
}

// ProbeClash probe clash service for httpUrl.
//
// Inspired by: https://github.com/MikeWang000000/ClashScan/
func ProbeClash(ctx context.Context, client *http.Client, httpUrl *url.URL) (*ClashInfo, error) {
	request := requestPool.Get().(*http.Request).WithContext(ctx)
	defer requestPool.Put(request)
	request.URL = httpUrl

	do := func(path string) (*http.Response, error) {
		request.URL.Path = path
		return client.Do(request)
	}

	resp, err := do("/")
	if err != nil {
		return nil, err
	}
	type clashHello struct {
		Hello string `json:"hello"`
	}
	hello := &clashHello{}
	err = json.NewDecoder(resp.Body).Decode(hello)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if hello.Hello != "clash" {
		return nil, E.New(httpUrl.Host, " seems not clash because hello is: ", hello.Hello)
	}

	resp, err = do("/version")
	if err != nil {
		return nil, E.Cause(err, "get version info")
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	type clashVersion struct {
		Version string `json:"version"`
		Premium bool   `json:"premium"`
		Meta    bool   `json:"meta"`
	}
	version := &clashVersion{}
	err = json.Unmarshal(content, version)
	if err != nil {
		return nil, E.Cause(err, "decode version info")
	}
	info := &ClashInfo{
		Type:    "Unknown",
		Version: "Unknown",
	}
	if rawVersion, ok := strings.CutPrefix(version.Version, "sing-box "); ok {
		info.Type = ClashTypeBox
		info.Version = rawVersion
	} else if version.Meta {
		info.Type = ClashTypeMihomo
		info.Version = version.Version
	} else if version.Premium {
		info.Type = ClashTypePremium
		info.Version = version.Version
	} else if version.Version != "" {
		info.Type = ClashTypeClassic
		info.Version = version.Version
	} else {
		info.guessVersion(do)
	}

	return info, nil
}

func (c *ClashInfo) guessVersion(do func(path string) (*http.Response, error)) {
	pathExists := func(path string) bool {
		resp, err := do(path)
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK, http.StatusUnauthorized:
			return true
		}
		return false
	}

	if pathExists("/memory") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.14.4 ~ Latest (guessed)"
		return
	} else if pathExists("/restart") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.14.3 (guessed)"
		return
	} else if pathExists("/dns") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.14.2 (guessed)"
		return
	} else if pathExists("/group") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.12.0 ~ v1.14.1 (guessed)"
		return
	} else if pathExists("/cache") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.10.0 ~ v1.11.8 (guessed)"
		return
	} else if pathExists("/script") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.9.1 (guessed)"
		return
	} else if pathExists("/providers/rules") {
		c.Type = ClashTypeMihomo
		c.Version = "v1.8.0 ~ v1.9.0 (guessed)"
		return
	} else if pathExists("/providers/proxies") {
		c.Type = ClashTypePremium //
		c.Version = "v0.17.0 ~ Latest (guessed)"
		return
	}
}
