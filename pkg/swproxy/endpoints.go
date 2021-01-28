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
		(strings.HasSuffix(ctx.Req.Host, "qpyou.cn") || strings.HasSuffix(ctx.Req.Host, "qpyou.cn:443"))

	if hostMatches {
		log.Trace().
			Str("log_type", "module").
			Str("module", "proxyGameEndpointMatcher").
			Str("host", ctx.Req.Host).
			Stringer("url", ctx.Req.URL).
			Str("method", ctx.Req.Method).
			Bool("endpoint_matches", methodMatches && hostMatches).
			Msg("Checking if endpoint matches")
	}

	return methodMatches && hostMatches
}

// Location Service Endpoint Matcher
type locationServiceEndpointMatcher struct{}

func newLocationServiceMatcher() *locationServiceEndpointMatcher {
	return new(locationServiceEndpointMatcher)
}

func (s locationServiceEndpointMatcher) HandleReq(_ *http.Request, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s locationServiceEndpointMatcher) HandleResp(_ *http.Response, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s locationServiceEndpointMatcher) matches(ctx *goproxy.ProxyCtx) bool {
	methodMatches := ctx.Req.Method == "GET"
	hostMatches := strings.HasPrefix(ctx.Req.Host, "summonerswar-") &&
		strings.HasSuffix(ctx.Req.Host, "qpyou.cn")
	urlMatches := ctx.Req.URL.Path == "/api/location_c2.php"

	if hostMatches {
		log.Trace().
			Str("log_type", "module").
			Str("module", "locationServiceEndpointMatcher").
			Str("host", ctx.Req.Host).
			Stringer("url", ctx.Req.URL).
			Str("method", ctx.Req.Method).
			Bool("endpoint_matches", methodMatches && hostMatches).
			Msg("Checking if endpoint matches")
	}

	return methodMatches && hostMatches && urlMatches
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
	hostMatches := strings.HasPrefix(ctx.Req.Host, "summonerswar-") &&
		strings.HasSuffix(ctx.Req.Host, "qpyou.cn")
	urlMatches := ctx.Req.URL.Path == "/api/gateway_c2.php"

	if hostMatches {
		log.Trace().
			Str("log_type", "module").
			Str("module", "gameEndpointMatcher").
			Str("host", ctx.Req.Host).
			Stringer("url", ctx.Req.URL).
			Str("method", ctx.Req.Method).
			Bool("endpoint_matches", methodMatches && hostMatches).
			Msg("Checking if endpoint matches")
	}

	return methodMatches && hostMatches && urlMatches
}

// Certificate endpoint matcher
// used to serve the certificate to the user if requested
type certificateEndpointMatcher struct{}

func newCertificateEndpointMatcher() *certificateEndpointMatcher {
	return new(certificateEndpointMatcher)
}

func (s certificateEndpointMatcher) HandleReq(_ *http.Request, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s certificateEndpointMatcher) HandleResp(_ *http.Response, ctx *goproxy.ProxyCtx) bool {
	return s.matches(ctx)
}

func (s certificateEndpointMatcher) matches(ctx *goproxy.ProxyCtx) bool {
	log.Debug().Msg("matched for user requested certificate")

	return ctx.Req.Method == "GET" && ctx.Req.URL.Path == "/ca.crt"
}
