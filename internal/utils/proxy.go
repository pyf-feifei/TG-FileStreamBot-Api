package utils

import (
	"EverythingSuckz/fsb/config"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// SetupProxy 配置HTTP代理（仅在开发环境中）
func SetupProxy(client *http.Client) {
	if config.ValueOf.Dev {
		proxyURL, _ := url.Parse("http://127.0.0.1:7890")
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
}

// CreateSOCKS5Dialer creates a SOCKS5 dialer for Telegram connections
// proxyURL format: socks5://127.0.0.1:1080 or socks5://user:pass@127.0.0.1:1080
func CreateSOCKS5Dialer(proxyURL string) (func(ctx context.Context, network, addr string) (net.Conn, error), error) {
	if proxyURL == "" {
		// Return default dialer if no proxy configured
		return func(ctx context.Context, network, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		}, nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	if u.Scheme != "socks5" {
		return nil, fmt.Errorf("unsupported proxy scheme: %s (only socks5 is supported)", u.Scheme)
	}

	// Extract auth credentials if present
	var auth *proxy.Auth
	if u.User != nil {
		password, _ := u.User.Password()
		auth = &proxy.Auth{
			User:     u.User.Username(),
			Password: password,
		}
	}

	// Remove user info from address
	proxyAddr := u.Host
	if proxyAddr == "" {
		return nil, fmt.Errorf("proxy address is empty")
	}

	// Create SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	// Return a context-aware dialer function
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// proxy.Dialer doesn't support context, but we can check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return dialer.Dial(network, addr)
	}, nil
}

// ValidateProxyURL validates if the proxy URL is in correct format
func ValidateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL format: %w", err)
	}

	if !strings.HasPrefix(u.Scheme, "socks5") {
		return fmt.Errorf("unsupported proxy scheme: %s (expected: socks5)", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("proxy host is empty")
	}

	return nil
}