/*
Copyright 2021 NDD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gnmiserver

import (
	"context"
	"net"
	"strconv"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/pkg/errors"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	"github.com/yndd/ndd-runtime/pkg/logging"
	targetv1 "github.com/yndd/target/apis/target/v1"
	"github.com/yndd/state/internal/collector"
	"github.com/yndd/state/internal/config"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// defaults
	defaultMaxSubscriptions = 64
	defaultMaxGetRPC        = 1024
)

// Option can be used to manipulate Options.
type Option func(GnmiServer)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) Option {
	return func(s GnmiServer) {
		s.WithLogger(log)
	}
}

func WithConfig(c config.Config) Option {
	return func(s GnmiServer) {
		s.WithConfig(c)
	}
}

func WithK8sClient(client client.Client) Option {
	return func(s GnmiServer) {
		s.WithK8sClient(client)
	}
}

type GnmiServer interface {
	WithLogger(log logging.Logger)
	WithK8sClient(client client.Client)
	WithConfig(c config.Config)
	Start() error
}

type serverConfig struct {
	// Address
	address string
	// Generic
	//maxSubscriptions int64
	//maxUnaryRPC      int64
	// TLS
	inSecure   bool
	skipVerify bool
	//caFile     string
	//certFile   string
	//keyFile    string
	// observability
	//debug         bool
}

type gnmiServerImpl struct {
	gnmi.UnimplementedGNMIServer

	cfg *serverConfig

	//k8s client
	client    client.Client
	newTarget func() targetv1.Tg
	// config per target
	config config.Config
	// state collector
	collector collector.Collector
	// gnmi calls
	subscribeRPCsem *semaphore.Weighted
	unaryRPCsem     *semaphore.Weighted
	// logging and parsing
	log logging.Logger
}

func New(ctx context.Context, opts ...Option) GnmiServer {
	//nn := func() targetv1.Tg { return &targetv1.Target{} }
	s := &gnmiServerImpl{
		cfg: &serverConfig{
			address:    ":" + strconv.Itoa(pkgmetav1.GnmiServerPort),
			skipVerify: true,
			inSecure:   true,
		},

		//newTarget: nn,
	}

	for _, opt := range opts {
		opt(s)
	}
	s.collector = collector.New(ctx,
		collector.WithLogger(s.log),
	)

	return s
}

func (s *gnmiServerImpl) WithLogger(log logging.Logger) {
	s.log = log
}

func (s *gnmiServerImpl) WithConfig(c config.Config) {
	s.config = c
}

func (s *gnmiServerImpl) WithK8sClient(client client.Client) {
	s.client = client
}

func (s *gnmiServerImpl) Start() error {
	log := s.log.WithValues("grpcServerAddress", s.cfg.address)
	log.Debug("grpc server run...")
	go func() {
		if err := s.run(); err != nil {
			log.Debug("failed to start gNMI server", "error", err)
		}
	}()
	return nil
}

// run GRPC Server
func (s *gnmiServerImpl) run() error {
	s.subscribeRPCsem = semaphore.NewWeighted(defaultMaxSubscriptions)
	s.unaryRPCsem = semaphore.NewWeighted(defaultMaxGetRPC)
	log := s.log.WithValues("grpcServerAddress", s.cfg.address)
	log.Debug("grpc server start...")

	// create a listener on a specific address:port
	l, err := net.Listen("tcp", s.cfg.address)
	if err != nil {
		return errors.Wrap(err, errCreateTcpListener)
	}

	// TODO, proper handling of the certificates with CERT Manager
	/*
		opts, err := s.serverOpts()
		if err != nil {
			return err
		}
	*/
	// create a gRPC server object
	grpcServer := grpc.NewServer()

	// attach the gnmi service to the grpc server
	gnmi.RegisterGNMIServer(grpcServer, s)

	// start the server
	log.Debug("grpc server serve...")
	if err := grpcServer.Serve(l); err != nil {
		s.log.Debug("Errors", "error", err)
		return errors.Wrap(err, errGrpcServer)
	}
	return nil
}
