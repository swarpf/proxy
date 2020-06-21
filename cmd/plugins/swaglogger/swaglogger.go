package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lyrex/swarpf/internal/proxyapiutil"
	pb "github.com/lyrex/swarpf/proto-gen/proxyapi"
)

func subscribedCommands() []string {
	return []string{"GetGuildWarBattleLogByWizardId", "GetGuildWarBattleLogByGuildId"}
}

func isSubscribedCommand(command string) bool {
	for _, b := range subscribedCommands() {
		if b == command {
			return true
		}
	}
	return false
}

func main() {
	// load configuration from command line or environment
	var (
		proxyAddr   = flag.String("proxyapi_addr", "127.0.0.1:8010", "Address of the proxy host")
		listenAddr  = flag.String("listen_addr", "127.0.0.1:11000", "Listen address for the plugin")
		development = flag.Bool("development", false, "Enable development logging")
	)
	flag.Parse()

	listenAddress := *listenAddr
	proxyAddress := *proxyAddr

	// setup logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *development {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	log.Logger = log.With().Timestamp().Str("log_type", "plugin").Str("plugin", "DebugOutput").Logger()

	// Main Program
	log.Info().
		Str("proxyAddr", proxyAddress).
		Msgf("Connecting SwagLogger plugin to proxy %s", proxyAddress)

	// initialize proxy consumer
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create listener")
	}

	log.Info().
		Str("listenAddr", listenAddress).
		Msgf("Listening for new proxy api connections on %s", listenAddress)

	s := grpc.NewServer()
	pb.RegisterProxyApiConsumerServer(s, &swagLoggerProxyApiConsumer{})

	go proxyapiutil.RegisterWithProxyApi(proxyAddress, listenAddress, subscribedCommands())

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Info().Str("reason", err.Error()).Msg("Server stopped listening")
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	proxyapiutil.DisconnectFromProxyApi(proxyAddress, listenAddress, subscribedCommands())

	log.Info().Err(err).Msg("SwagLogger plugin ended")
}

// proxy API consumer
type swagLoggerProxyApiConsumer struct {
	pb.UnimplementedProxyApiConsumerServer
}

func (s *swagLoggerProxyApiConsumer) OnReceiveApiEvent(_ context.Context, ev *pb.ApiEvent) (*empty.Empty, error) {
	if !isSubscribedCommand(ev.Command) {
		return &empty.Empty{}, nil
	}

	requestContent := map[string]interface{}{}
	if err := json.Unmarshal([]byte(ev.Request), &requestContent); err != nil {
		log.Error().Err(err).Msg("Failed to deserializie SWAG request")
		return &empty.Empty{}, errors.New("error while deserializing SWAG request")
	}

	responseContent := map[string]interface{}{}
	if err := json.Unmarshal([]byte(ev.Response), &responseContent); err != nil {
		log.Error().Err(err).Msg("Failed to deserializie SWAG response")
		return &empty.Empty{}, errors.New("error while deserializing SWAG response")
	}

	command := responseContent["command"].(string)
	wizardId := requestContent["wizard_id"].(float64)

	log.Info().
		Str("command", command).
		Float64("wizard_id", wizardId).
		Msg("Uploading guild war data to SWAG...")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(ev.Response).
		Post("https://gw.swop.one/data/upload/")

	if err != nil {
		log.Error().Err(err).
			Str("command", command).
			Float64("wizardId", wizardId).
			Msg("SWAG upload failed")
		return &empty.Empty{}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		log.Error().
			Str("command", command).
			Float64("wizardId", wizardId).
			Int("StatusCode", resp.StatusCode()).
			Msgf("SWAG upload failed. Status %d", resp.StatusCode())
		return &empty.Empty{}, nil
	}

	log.Info().
		Str("command", command).
		Float64("wizardId", wizardId).
		Msg("SWAG upload successful.")

	return &empty.Empty{}, nil
}
