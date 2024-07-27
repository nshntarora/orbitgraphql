package balancer

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// DefaultProxyBalancer is a simple proxy balancer that always proxies to a single target instead of load balancing
type DefaultProxyBalancer struct {
	target *middleware.ProxyTarget
}

func NewDefaultProxyBalancer(target *middleware.ProxyTarget) middleware.ProxyBalancer {
	return &DefaultProxyBalancer{
		target: target,
	}
}

func (b *DefaultProxyBalancer) AddTarget(target *middleware.ProxyTarget) bool {
	b.target = target
	return true
}

func (b *DefaultProxyBalancer) RemoveTarget(targetURL string) bool {
	b.target = nil
	return false
}

func (b *DefaultProxyBalancer) Next(echo.Context) *middleware.ProxyTarget {
	if b.target != nil {
		return b.target
	}
	return nil
}
