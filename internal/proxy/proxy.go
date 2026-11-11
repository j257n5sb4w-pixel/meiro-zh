// Package proxy routes Meiro's network traffic through a proxy the user sets,
// either in the sign-in window or in Settings.
//
// It covers every way the app reaches the network: Go's own HTTP clients (which
// take their proxy from Proxy), the ffmpeg, yt-dlp and browser processes Meiro
// starts (which read the HTTP_PROXY family from their environment), and the
// sign-in browser window (which needs --proxy-server).
package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	// checkURL is reached by Check: a host the app needs anyway, so a working
	// answer means the rest of the app can reach YouTube too.
	checkURL     = "https://music.youtube.com/"
	checkTimeout = 15 * time.Second
)

// loopback never goes through the proxy: local artwork, the audio cache and the
// loopback debugging port of the sign-in window all live there.
var loopback = []string{"localhost", "127.0.0.1", "::1"}

var (
	mu      sync.RWMutex
	address string
	parsed  *url.URL
)

// Set records the proxy address, in any of the forms a user is likely to paste:
// "127.0.0.1:7890", "http://127.0.0.1:7890", "socks5://127.0.0.1:1080". An
// empty value turns the proxy off.
func Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		Clear()
		return nil
	}
	target, err := parse(value)
	if err != nil {
		return err
	}
	mu.Lock()
	address = value
	parsed = target
	mu.Unlock()
	export(target)
	return nil
}

// Clear turns the proxy off.
func Clear() {
	mu.Lock()
	address = ""
	parsed = nil
	mu.Unlock()
	unexport()
}

// Address reports the address the user set, empty when the proxy is off.
func Address() string {
	mu.RLock()
	defer mu.RUnlock()
	return address
}

// On reports whether a proxy is set.
func On() bool { return Address() != "" }

// Proxy decides where one request goes. It is the http.Transport hook, so it
// runs on every connection.
func Proxy(req *http.Request) (*url.URL, error) {
	mu.RLock()
	target := parsed
	mu.RUnlock()
	if target == nil {
		return http.ProxyFromEnvironment(req)
	}
	if req != nil && req.URL != nil && isLoopback(req.URL.Hostname()) {
		return nil, nil
	}
	return target, nil
}

// Transport builds an HTTP transport that uses the proxy. Callers keep their own
// timeouts; only the proxy decision is shared.
func Transport() *http.Transport {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	transport := base.Clone()
	transport.Proxy = Proxy
	return transport
}

// Install makes the default transport take the proxy, so every client that
// keeps http.DefaultTransport follows it without being handed a transport of
// its own: the audio cache, the lyrics and Last.fm clients, and the covers.
func Install() {
	http.DefaultTransport = Transport()
}

// Apply points a command at the proxy, so ffmpeg, yt-dlp and the browser Meiro
// starts reach it too. It leaves the environment alone when the proxy is off.
func Apply(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	mu.RLock()
	target := parsed
	mu.RUnlock()
	if target == nil {
		return
	}
	env := cmd.Env
	if env == nil {
		env = os.Environ()
	}
	cmd.Env = withProxy(env, target)
}

// ServerArg is the value Chromium-based browsers want in --proxy-server.
func ServerArg() string {
	mu.RLock()
	target := parsed
	mu.RUnlock()
	if target == nil {
		return ""
	}
	return target.String()
}

// Check reaches checkURL through the proxy and reports what went wrong, so the
// user learns whether an address works before signing in through it.
func Check() error {
	client := &http.Client{Transport: Transport(), Timeout: checkTimeout}
	req, err := http.NewRequest(http.MethodHead, checkURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("%s answered %s", checkURL, resp.Status)
	}
	return nil
}

func parse(value string) (*url.URL, error) {
	text := value
	if !strings.Contains(text, "://") {
		text = "http://" + text
	}
	target, err := url.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("that is not a proxy address: %w", err)
	}
	switch strings.ToLower(target.Scheme) {
	case "http", "https", "socks5", "socks5h", "socks4":
	default:
		return nil, fmt.Errorf("%q is not a proxy type Meiro knows: use http, https or socks5", target.Scheme)
	}
	if target.Host == "" {
		return nil, fmt.Errorf("that is not a proxy address: the host is missing")
	}
	if path := strings.TrimSuffix(target.Path, "/"); path != "" {
		return nil, fmt.Errorf("that is not a proxy address: leave the path off")
	}
	target.Path = ""
	return target, nil
}

func isLoopback(host string) bool {
	host = strings.ToLower(strings.Trim(host, "[]"))
	for _, name := range loopback {
		if host == name {
			return true
		}
	}
	return false
}

// withProxy rewrites the proxy variables in env and keeps loopback direct.
func withProxy(env []string, target *url.URL) []string {
	value := target.String()
	set := map[string]string{
		"HTTP_PROXY":  value,
		"HTTPS_PROXY": value,
		"ALL_PROXY":   value,
		"http_proxy":  value,
		"https_proxy": value,
		"all_proxy":   value,
		"NO_PROXY":    strings.Join(loopback, ","),
	}
	out := make([]string, 0, len(env)+len(set))
	for _, entry := range env {
		name, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if _, replaced := set[name]; replaced {
			continue
		}
		out = append(out, entry)
	}
	for name, value := range set {
		out = append(out, name+"="+value)
	}
	return out
}

// export mirrors the proxy into this process, so Go subprocesses and mygo's own
// fetches follow it as well.
func export(target *url.URL) {
	for _, entry := range withProxy(os.Environ(), target) {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		_ = os.Setenv(name, value)
	}
}

func unexport() {
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy", "NO_PROXY", "no_proxy"} {
		_ = os.Unsetenv(name)
	}
}
