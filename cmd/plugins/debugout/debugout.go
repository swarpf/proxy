package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lyrex/swarpf/internal/proxyapiutil"
	pb "github.com/lyrex/swarpf/proto-gen/proxyapi"
)

func subscribedCommands() []string {
	return []string{"*"}
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
		Msgf("Connecting DebugOutput plugin to proxy %s", proxyAddress)

	// initialize proxy consumer
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create listener")
	}

	log.Info().
		Str("listenAddr", listenAddress).
		Msgf("Listening for new proxy api connections on %s", listenAddress)

	s := grpc.NewServer()
	pb.RegisterProxyApiConsumerServer(s, &debugOutputProxyApiConsumer{})

	go proxyapiutil.RegisterWithProxyApi(proxyAddress, listenAddress, subscribedCommands())

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Info().Str("reason", err.Error()).Msg("Server stopped listening")
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	log.Info().Err(err).Msg("DebugOutput plugin ended")

	go proxyapiutil.DisconnectFromProxyApi(proxyAddress, listenAddress, subscribedCommands())
}

// proxy API consumer
type debugOutputProxyApiConsumer struct {
	pb.UnimplementedProxyApiConsumerServer
}

func (s *debugOutputProxyApiConsumer) OnReceiveApiEvent(_ context.Context, ev *pb.ApiEvent) (*empty.Empty, error) {
	log.Debug().Timestamp().
		Str("command", ev.GetCommand()).
		Str("request", ev.GetRequest()).
		Str("response", ev.GetResponse()).
		Msg("Debug Output Plugin")

	return &empty.Empty{}, nil
}
