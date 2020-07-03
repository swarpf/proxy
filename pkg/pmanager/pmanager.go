package pmanager

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/swarpf/proxy/pkg/apiemitter"
	"github.com/swarpf/proxy/pkg/events"
	pb "github.com/swarpf/proxy/proto-gen/proxyapi"
)

type proxyConsumer struct {
	Commands []string
	Client   pb.ProxyApiConsumerClient
}

var activeProxyConsumers map[string]proxyConsumer
var proxyApiLogger zerolog.Logger

type ProxyManager struct {
	em *apiemitter.Emitter
}

func NewProxyManager(proxyApiAddr string) *ProxyManager {
	activeProxyConsumers = make(map[string]proxyConsumer)
	proxyApiLogger = log.With().Timestamp().Str("log_type", "module").Str("module", "ProxyAPI").Logger()

	go func() {
		// initialize proxy consumer
		lis, err := net.Listen("tcp", proxyApiAddr)
		if err != nil {
			proxyApiLogger.Fatal().Err(err).Msg("failed to create listener")
		}

		s := grpc.NewServer()
		pb.RegisterProxyApiServer(s, &proxyApiServer{})

		proxyApiLogger.Info().
			Str("proxyApiAddr", proxyApiAddr).
			Msgf("Listening for new connections at %s", proxyApiAddr)

		err = s.Serve(lis)
		proxyApiLogger.Info().Err(err).Msg("stopped listening for new proxy api connections")
	}()

	return &ProxyManager{em: apiemitter.New(1)}
}

func (pm *ProxyManager) Publish(topic string, msg events.ApiEventMsg) {
	go pm.em.Emit(topic, msg)

	for consumerAddr, consumer := range activeProxyConsumers {
		for _, command := range consumer.Commands {
			if matched, err := path.Match(command, msg.Command); err != nil {
				proxyApiLogger.Error().Err(err).
					Str("command", command).
					Str("msg.Command", msg.Command).
					Msg("Failed to match command")
			} else if !matched {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err := consumer.Client.OnReceiveApiEvent(ctx,
				&pb.ApiEvent{Command: msg.Command, Request: msg.Request, Response: msg.Response})

			if err != nil {
				proxyApiLogger.Error().Err(err).Str("consumerAddr", consumerAddr).
					Msg("failed to publish api event")
				cancel()
				continue
			}

			proxyApiLogger.Debug().
				Str("consumerAddr", consumerAddr).
				Str("command", command).
				Str("msg.Command", msg.Command).
				Msgf("Published %s to Proxy API consumer at %s", msg.Command, consumerAddr)

			cancel()
		}
	}
}

func (pm *ProxyManager) Subscribe(topic string) <-chan apiemitter.Event {
	return pm.em.On(topic)
}

func (pm *ProxyManager) Unsubscribe(topic string, ch ...<-chan apiemitter.Event) {
	pm.em.Off(topic, ch...)
}

func (pm *ProxyManager) Shutdown() {
	pm.em.Off("*")
}

// proxy api provider server
// ProxyApiProvider server implementation
type proxyApiServer struct {
	pb.UnimplementedProxyApiServer
}

func (s *proxyApiServer) Register(_ context.Context, opts *pb.ProxyApiOptions) (*pb.ProxyApiProviderResponse, error) {
	proxyApiLogger.Info().
		Str("consumerAddr", opts.Address).
		Strs("commands", opts.Commands).
		Msg("New request to register a proxy api consumer")

	_, exists := activeProxyConsumers[opts.Address]
	if exists {
		proxyApiLogger.Warn().Str("remoteAddr", opts.Address).
			Msg("Proxy api client with this address already exists")

		err := errors.New("proxy api client with this address already exists")
		return &pb.ProxyApiProviderResponse{Success: false, Error: err.Error()}, err
	}

	conn, err := grpc.Dial(opts.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		proxyApiLogger.Error().Err(err).Str("remoteAddr", opts.Address).Msg("did not connect")
		return nil, fmt.Errorf("failed to connect to %s", opts.Address)
	}
	// defer func() {
	// 	if e := conn.Close(); e != nil {
	// 		proxyApiLogger.Error().Err(e).Str("remoteAddr", opts.Address).Msg("failed to close connection")
	// 	}
	// }()

	c := pb.NewProxyApiConsumerClient(conn)
	activeProxyConsumers[opts.Address] = proxyConsumer{
		Commands: opts.Commands,
		Client:   c,
	}

	proxyApiLogger.Info().
		Str("consumerAddr", opts.Address).
		Strs("commands", opts.Commands).
		Msg("Successfully registered a proxy api consumer")

	return &pb.ProxyApiProviderResponse{Success: true}, nil
}

func (s *proxyApiServer) Disconnect(_ context.Context, opts *pb.ProxyApiOptions) (*pb.ProxyApiProviderResponse, error) {
	proxyApiLogger.Info().
		Str("consumerAddr", opts.Address).
		Strs("commands", opts.Commands).
		Msg("New request to disconnect a proxy api consumer")

	_, exists := activeProxyConsumers[opts.Address]
	if !exists {
		proxyApiLogger.Warn().Str("remoteAddr", opts.Address).
			Msg("proxy api client with this does not exists")

		err := errors.New("proxy api client with this address does not exists")
		return &pb.ProxyApiProviderResponse{Success: false, Error: err.Error()}, err
	}

	delete(activeProxyConsumers, opts.Address)

	proxyApiLogger.Info().
		Str("consumerAddr", opts.Address).
		Strs("commands", opts.Commands).
		Msg("Successfully disconnected a proxy api consumer")

	return &pb.ProxyApiProviderResponse{Success: true}, nil
}
