package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lyrex/swarpf/pkg/events"
	"github.com/lyrex/swarpf/pkg/pmanager"
	"github.com/lyrex/swarpf/pkg/swproxy"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// todo(lyrex): Move this into some kind of configuration
	const host = "0.0.0.0"
	const port = 8009

	mainLogger := log.With().Timestamp().Str("log_type", "app").Str("module", "main").Logger()
	mainLogger.Info().
		Str("host", host).
		Uint16("port", port).
		Msgf("Server listening to %s:%d", host, port)

	apiEvents := make(chan events.ApiEventMsg, 1)

	// initialize proxy manager
	pm := pmanager.NewProxyManager()

	// initialize proxy
	swProxy := swproxy.New(apiEvents, swproxy.ProxyConfiguration{
		CertificateDirectory: "./cert/",
		InterceptHttps:       false,
	})
	httpProxy := swProxy.CreateProxy()

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: httpProxy}
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
	stop := make(chan os.Signal)
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
