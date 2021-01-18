package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/swarpf/proxy/pkg/events"
	"github.com/swarpf/proxy/pkg/pmanager"
	"github.com/swarpf/proxy/pkg/swproxy"
)

func main() {
	// load configuration from command line or environment
	var (
		listenAddr     = flag.String("proxy_listen_addr", "0.0.0.0:8010", "Listen address for the http proxy")
		proxyApiAddr   = flag.String("proxyapi_listen_addr", "0.0.0.0:11000", "Listen address for the proxy API")
		development    = flag.Bool("development", false, "Enable development logging")
		interceptHttps = flag.Bool("intercept_https", false, "Enable HTTPS interception")
	)
	flag.Parse()

	listenAddress := *listenAddr
	proxyApiAddress := *proxyApiAddr

	// setup logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *development {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	log.Logger = log.With().Timestamp().Str("log_type", "app").Str("app", "Proxy").Logger()

	mainLogger := log.With().Str("module", "main").Logger()
	mainLogger.Info().
		Str("listenAddress", listenAddress).
		Msgf("Server listening to %s", listenAddress)

	apiEvents := make(chan events.ApiEventMsg, 1)

	// initialize proxy manager
	pm := pmanager.NewProxyManager(proxyApiAddress)

	// initialize proxy
	swProxy := swproxy.New(apiEvents, swproxy.ProxyConfiguration{
		CertificateDirectory: "./cert/",
		InterceptHttps:       *interceptHttps,
	})
	httpProxy := swProxy.CreateProxy()

	server := &http.Server{Addr: listenAddress, Handler: httpProxy}
	go func() {
		err := server.ListenAndServe()

		if err != nil {
			mainLogger.Info().
				Timestamp().
				Str("log_type", "app").
				Str("module", "main").
				Str("reason", err.Error()).
				Msg("Proxy stopped listening")
		}
	}()

	// process api events
	go sendCommandsToProxyManager(pm, apiEvents)

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	// shutdown communitcation
	pm.Shutdown()
	close(apiEvents)

	mainLogger.Info().Msg("Shutting down proxy...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		mainLogger.Panic().Err(err).Msg("")
	}

	mainLogger.Info().Msg("Proxy shut down")
}

func sendCommandsToProxyManager(pm *pmanager.ProxyManager, ev chan events.ApiEventMsg) {
	for apiEvent := range ev {
		requestContent := map[string]interface{}{}
		if err := json.Unmarshal([]byte(apiEvent.Request), &requestContent); err != nil {
			log.Error().Err(err).Msg("Error while deserializing API request")
			continue
		}
		command, ok := requestContent["command"].(string)
		if !ok {
			log.Error().Str("request", apiEvent.Request).Msg("Failed to extract command from request")
			continue
		}

		apiEvent.Command = command
		pm.Publish(command, apiEvent)
	}
}
