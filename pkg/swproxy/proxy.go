package swproxy

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	grpczerolog "github.com/cheapRoc/grpc-zerolog"
	"github.com/elazarl/goproxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/swarpf/proxy/pkg/events"
	"github.com/swarpf/proxy/pkg/utils"
)

type ProxyConfiguration struct {
	CertificateDirectory string `default:"./certs/"`
	InterceptHttps       bool
	ForceHttpDowngrade   bool `default:"false"`
	Verbose              bool `default:"false"`
}

type Proxy struct {
	log           zerolog.Logger
	eventChan     chan events.ApiEventMsg
	configuration ProxyConfiguration
}

// proxy.New : Create a new proxy instance for further use
func New(ev chan events.ApiEventMsg, configuration ProxyConfiguration) *Proxy {
	if ev == nil {
		log.Panic().Msg("ev is not a valid ApiEventMsg channel")
		return nil
	}

	return &Proxy{
		log:           log.With().Timestamp().Str("log_type", "module").Str("module", "Proxy").Logger(),
		eventChan:     ev,
		configuration: configuration,
	}
}

func (p *Proxy) CreateProxy() http.Handler {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = grpczerolog.New(log.Logger) // todo(lyrex): this need some kind of better implementation that does not just throw everything into INFO
	proxy.Verbose = p.configuration.Verbose

	if p.configuration.ForceHttpDowngrade {
		p.log.Info().Msg("HTTPS -> HTTP downgrade is enabled")

		// match the /api/location_c2.php endpoint and modify the body if necessary
		proxy.OnResponse(newLocationServiceMatcher()).
			DoFunc(p.onLocationResponse)
	}

	if p.configuration.InterceptHttps {
		p.log.Info().Msg("HTTPS interception is enabled")

		rootCa := getRootCA(p.configuration.CertificateDirectory)
		if err := setCA(rootCa); err != nil {
			p.log.Fatal().Err(err).Msg("could not set proxy CA")
			return nil
		}

		proxy.OnRequest(newProxyGameEndpointMatcher()).HandleConnect(goproxy.AlwaysMitm)

		proxy.OnRequest(newCertificateEndpointMatcher()).DoFunc(
			func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				p.log.Debug().Msg("user requested certificate")
				return req,
					goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusOK, string(rootCa.Certificate[0]))
			})
	}

	proxy.OnRequest(newGameEndpointMatcher()).
		DoFunc(p.onRequest)

	proxy.OnResponse(newGameEndpointMatcher()).
		DoFunc(p.onResponse)

	return proxy
}

func (p *Proxy) onRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	requestLogger := p.log.With().Int64("ctx.Session", ctx.Session).Logger()

	requestLogger.Trace().
		Stringer("ctx.Req.URL", ctx.Req.URL).
		Interface("ctx.Req.Header", ctx.Req.Header).
		Msg("New outgoing request")

	if req == nil || req.ContentLength == 0 || req.Body == nil {
		requestLogger.Info().Msg("Sending empty request to API")
		return req, nil
	}
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		requestLogger.Error().Err(err).Msg("could not read request body")
		return req, nil
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	reqContent := string(reqBody[:])
	plainContent, err := p.readBody(reqContent, false)
	if err != nil {
		// do not log here since we're logging the actual error in readBody
		return req, nil
	}

	requestLogger.Trace().
		Str("encryptedContent", reqContent).
		Str("plainContent", plainContent).
		Msg("Sending request from API")

	ctx.UserData = plainContent

	return req, nil
}

func (p *Proxy) onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	responseLogger := p.log.With().Int64("ctx.Session", ctx.Session).Logger()

	responseLogger.Trace().
		Stringer("ctx.Req.URL", ctx.Req.URL).
		Interface("ctx.Req.Header", ctx.Req.Header).
		Msg("New incoming response")

	if resp == nil || resp.ContentLength == 0 || resp.Body == nil {
		responseLogger.Info().Msg("Received empty reponse from API")
		return resp
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		responseLogger.Error().Err(err).Msg("could not read response body")
		return resp
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

	respContent := string(respBody[:])
	responsePlainContent, err := p.readBody(respContent, true)
	if err != nil {
		// do not log here since we're logging the actual error in readBody
		return resp
	}

	responseLogger.Trace().
		Str("encryptedContent", respContent).
		Str("plainContent", responsePlainContent).
		Msg("Receiving response from API")

	requestPlainContent := ctx.UserData.(string)

	// send ApiEvent to event message
	p.eventChan <- events.ApiEventMsg{
		Request:  requestPlainContent,
		Response: responsePlainContent,
	}

	return resp
}

func (p *Proxy) onLocationResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	responseLogger := p.log.With().
		Int64("ctx.Session", ctx.Session).
		Str("tag", "location_endpoint").
		Logger()

	responseLogger.Trace().
		Stringer("ctx.Req.URL", ctx.Req.URL).
		Interface("ctx.Req.Header", ctx.Req.Header).
		Msg("New incoming location response")

	if resp == nil || resp.ContentLength == 0 || resp.Body == nil {
		responseLogger.Info().Msg("Received empty reponse from location API")
		return resp
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		responseLogger.Error().Err(err).Msg("could not read location response body")
		return resp
	}
	// NOTE: we keep this here to not break the buffer if we need to bail early
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	responseText, err := p.readBody(string(bodyBytes), true)
	if err != nil {
		// do not log here since we're logging the actual error in readBody
		return resp
	}

	responseLogger.Info().
		Str("plainContent", responseText).
		Msg("Receiving location response from API")

	// ensure we really do have a correctly decryped body
	if !strings.Contains(responseText, "server_url_list") {
		p.log.Warn().Msg("Location API response does not contain server url list.")
		return resp
	}

	// replace all occurences of https with http
	modifiedResponseText := strings.ReplaceAll(responseText, "https:", "http:")
	modifiedResponseText = strings.ReplaceAll(modifiedResponseText, "HTTPS:", "HTTP:")

	// encrypt and compress new response text
	workmem, err := utils.CompressBytes([]byte(modifiedResponseText)) // TODO: maybe do something similiar to Encoding.ASCII.GetBytes
	if err != nil {
		p.log.Warn().Err(err).Msg("could not compress data")
		return resp
	}

	workmem, err = utils.EncryptBytes(workmem)
	if err != nil {
		p.log.Warn().Err(err).Msg("could not encrypt bytes")
		return resp
	}

	modifiedBodyBytes := []byte(base64.StdEncoding.EncodeToString(workmem))

	// write modified body to response
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
	resp.ContentLength = int64(len(modifiedBodyBytes))

	responseLogger.Info().
		Bytes("encryptedContent", modifiedBodyBytes).
		Str("plainContent", modifiedResponseText).
		Msg("Modified server response and replaced HTTPS with HTTP.")

	return resp
}

func (p *Proxy) readBody(body string, decompress bool) (string, error) {
	if len(body) == 0 {
		return "", nil
	}

	encryptedBody := body

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedBody)
	if err != nil {
		p.log.Error().Err(err).Msg("could not decode content")
		return "", errors.New("could not decode body content")
	}

	decryptedBytes, err := utils.DecryptBytes(encryptedBytes)
	if err != nil {
		p.log.Error().Err(err).Msg("could not decrypt data")
		return "", errors.New("could not decrypt body data")
	}

	// we're done if we don't need to decompress any data
	if !decompress {
		return string(decryptedBytes[:]), nil
	}

	// otherwise decompress and return decompressed data
	decompressedBytes, err := utils.DecompressBytes(decryptedBytes)
	if err != nil {
		p.log.Error().Err(err).Msg("could not decompress data")
		return "", errors.New("could not decompress body data")
	}

	return string(decompressedBytes[:]), nil
}
