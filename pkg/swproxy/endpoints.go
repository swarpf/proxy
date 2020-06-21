package swproxy

import (
	"github.com/elazarl/goproxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"net/http"
	"strings"
)

// Proxy Game Endpoint Matcher
// used to determine if there's a CONNECT request to the Com2uS game server
type proxyGameEndpointMatcher struct {
	Log zerolog.Logger
}

func newProxyGameEndpointMatcher() *proxyGameEndpointMatcher {
	return new(proxyGameEndpointMatcher)
}

func (s proxyGameEndpointMatcher) HandleReq(_ *http.Request, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s proxyGameEndpointMatcher) HandleResp(_ *http.Response, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s proxyGameEndpointMatcher) matches(ctx *goproxy.ProxyCtx) bool {
	methodMatches := ctx.Req.Method == "CONNECT"

	hostMatches := strings.HasPrefix(ctx.Req.Host, "summonerswar-") &&
		strings.HasSuffix(ctx.Req.Host, "qpyou.cn")

	if hostMatches {
		log.Trace().
			Str("log_type", "module").
			Str("module", "ProxyGameEndpointMatcher").
			Str("host", ctx.Req.Host).
			Stringer("url", ctx.Req.URL).
			Str("method", ctx.Req.Method).
			Bool("isGetMethod", methodMatches).
			Bool("hostMatches", hostMatches).
			Msg("")
	}

	return methodMatches && hostMatches
}

// // Game Endpoint Matcher
// used to intercept requests from and to the Com2uS game server
type gameEndpointMatcher struct{}

func newGameEndpointMatcher() *gameEndpointMatcher {
	return new(gameEndpointMatcher)
}

func (s gameEndpointMatcher) HandleReq(_ *http.Request, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s gameEndpointMatcher) HandleResp(_ *http.Response, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s gameEndpointMatcher) matches(ctx *goproxy.ProxyCtx) bool {
	methodMatches := ctx.Req.Method == "GET" || ctx.Req.Method == "POST"
	hostMatches := strings.Contains(ctx.Req.Host, "qpyou.cn")
	urlMatches := ctx.Req.URL.Path == "/api/gateway_c2.php"

	log.Trace().
		Str("log_type", "module").
		Str("module", "GameEndpointMatcher").
		Str("host", ctx.Req.Host).
		Stringer("url", ctx.Req.URL).
		Str("method", ctx.Req.Method).
		Bool("isGetMethod", methodMatches).
		Bool("hostMatches", hostMatches).
		Bool("urlMatches", hostMatches).
		Msg("")

	return methodMatches && hostMatches && urlMatches
}
