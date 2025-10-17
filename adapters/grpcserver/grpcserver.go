package grpcmanager

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/abhissng/neuron/adapters/jwt"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
	"github.com/abhissng/neuron/utils/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// ServerConfig holds gRPC server configurations
type ServerConfig struct {
	port           int
	certFile       string
	keyFile        string
	caFile         string
	jwtSecret      string
	enableMetrics  bool
	serviceName    string
	maxRecvMsgSize int // In megabytes
	log            *log.Log
}

// Option is a function that modifies ServerConfig
type Option func(*ServerConfig)

// WithPort sets the gRPC server port
func WithPort(port int) Option {
	return func(c *ServerConfig) {
		c.port = port
	}
}

// WithTLS enables TLS with provided cert/key
func WithTLS(certFile, keyFile, caFile string) Option {
	return func(c *ServerConfig) {
		c.certFile = certFile
		c.keyFile = keyFile
		c.caFile = caFile
	}
}

// WithJWT enables authentication using JWT secret
func WithJWT(secret string) Option {
	return func(c *ServerConfig) {
		c.jwtSecret = secret
	}
}

// WithMetrics enables Prometheus monitoring
func WithMetrics() Option {
	return func(c *ServerConfig) {
		c.enableMetrics = true
	}
}

// WithMaxRecvMsgSize sets max received message size (MB)
func WithMaxRecvMsgSize(size int) Option {
	return func(c *ServerConfig) {
		c.maxRecvMsgSize = size
	}
}

func WithLogger(log *log.Log) Option {
	return func(c *ServerConfig) {
		c.log = log
	}
}

// Server represents a gRPC server
type Server struct {
	server   *grpc.Server
	config   ServerConfig
	registry *prometheus.Registry
}

// NewServer creates a new gRPC server with option pattern
func NewServer(opts ...Option) (*Server, error) {
	// Default configuration
	config := ServerConfig{
		port:           50051,
		serviceName:    "default-service",
		maxRecvMsgSize: 4, // Default 4MB
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}
	if config.log == nil {
		config.log = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
		config.log.Warn("Logger not provided, using default logger")
	}

	// Setup gRPC options
	grpcOpts := []grpc.ServerOption{}

	// TLS Configuration
	tlsConfig, err := createTLSConfig(config)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	// Interceptors (Middleware)
	unaryInterceptors, streamInterceptors := buildInterceptors(config)
	grpcOpts = append(grpcOpts,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
		grpc.MaxRecvMsgSize(config.maxRecvMsgSize*1024*1024),
	)

	// Create gRPC Server
	s := &Server{
		server:   grpc.NewServer(grpcOpts...),
		config:   config,
		registry: prometheus.NewRegistry(),
	}

	return s, nil
}

// createTLSConfig sets up TLS for gRPC server
func createTLSConfig(config ServerConfig) (*tls.Config, error) {
	if config.certFile == "" || config.keyFile == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(config.certFile, config.keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Mutual TLS (mTLS)
	if config.caFile != "" {
		caCert, err := os.ReadFile(config.caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %v", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to add CA cert")
		}
		tlsConfig.ClientCAs = caPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// buildInterceptors sets up gRPC middlewares
func buildInterceptors(config ServerConfig) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	var unary []grpc.UnaryServerInterceptor
	var stream []grpc.StreamServerInterceptor

	// Logging Interceptor
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	unary = append(unary, logging.UnaryServerInterceptor(InterceptorLogger(config.log), loggingOpts...))
	stream = append(stream, logging.StreamServerInterceptor(InterceptorLogger(config.log), loggingOpts...))

	// Authentication
	if config.jwtSecret != "" {
		authFunc := createAuthFunc(config.jwtSecret)
		unary = append(unary, auth.UnaryServerInterceptor(authFunc))
		stream = append(stream, auth.StreamServerInterceptor(authFunc))
	}

	// Metrics
	if config.enableMetrics {
		grpc_prometheus.EnableHandlingTimeHistogram()
		unary = append(unary, grpc_prometheus.UnaryServerInterceptor)
		stream = append(stream, grpc_prometheus.StreamServerInterceptor)
	}

	// Recovery (Panic Handling)
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(recoveryHandler),
	}
	unary = append(unary, recovery.UnaryServerInterceptor(recoveryOpts...))
	stream = append(stream, recovery.StreamServerInterceptor(recoveryOpts...))
	return unary, stream
}

// InterceptorLogger is a simple logging manager
func InterceptorLogger(l *log.Log) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		requestID, _ := ctx.Value(constant.RequestID).(types.StringConstant)
		l.Printf(zapcore.Level(lvl), "[%s] %s: %v", requestID, msg, fields) // #nosec G115
	})
}

// recoveryHandler handles panics in gRPC calls
func recoveryHandler(p interface{}) error {
	helpers.Println(constant.ERROR, "panic recovered: ", p)
	return status.Errorf(codes.Internal, "internal server error")
}

// createAuthFunc sets up authentication logic
func createAuthFunc(secret string) auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		// Validate JWT token (Add your own logic)
		claims, err := jwt.ValidateJWT(token, secret, []string{})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, types.StringConstant(constant.Service), claims.ServiceName)
		ctx = context.WithValue(ctx, types.StringConstant(constant.Roles), claims.Roles)
		ctx = context.WithValue(ctx, types.StringConstant(constant.RequestID), random.GenerateUUID())

		return ctx, nil
	}
}

// Start the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s.config.log.Info(fmt.Sprintf("Starting gRPC server on port %d", s.config.port))

	return s.server.Serve(lis)
}

// GracefulStop stops the gRPC server gracefully
func (s *Server) GracefulStop() {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ServerDefaultGracefulTime)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
	case <-stopped:
	}
}

// Will remove the commented code

/*
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterDiscoveryServiceServer(s, &server{})
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}




type server struct {
	pb.UnimplementedDiscoveryServiceServer
}

func (s *server) ProcessDiscovery(ctx context.Context, req *pb.DiscoveryMessage) (*pb.DiscoveryMessage, error) {
	// Unpack core payload
	var core pb.Core
	if err := req.Payload.UnmarshalTo(&core); err != nil {
		return nil, fmt.Errorf("failed to unpack payload: %v", err)
	}

	// Process the core data
	log.Printf("Processing request from service: %s", req.CurrentService)
	log.Printf("Core data: %+v", core)

	// Create response
	resPayload, _ := anypb.New(&pb.Core{
		User:    core.User,
		Primary: core.Primary,
		Input:   core.Input,
	})

	return &pb.DiscoveryMessage{
		CorrelationId:  req.CorrelationId,
		RequestId:      req.RequestId,
		Payload:        resPayload,
		Status:         pb.DiscoveryMessage_STATUS_COMPLETED,
		Timestamp:      timestamppb.New(time.Now()),
		CurrentService: "discovery-service",
	}, nil
}

/*
		conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
		if err != nil {
			fmt.Printf("did not connect: %v", err)
		}
		defer conn.Close()

		c := pb.NewDiscoveryServiceClient(conn)

		// Create core payload
		core := &pb.Core{
			User: &pb.Core_UserInformation{
				UserId: "user-123",
				Email:  "user@example.com",
			},
			Primary: &pb.Core_PrimaryInfo{
				PrimaryId:   "primary-456",
				PrimaryType: "default",
			},
		}

		payload, _ := anypb.New(core)

		// Create request
		reqg := &pb.DiscoveryMessage{
			CorrelationId:  "corr-001",
			RequestId:      "req-002",
			Payload:        payload,
			Status:         pb.DiscoveryMessage_STATUS_PENDING,
			Action:         pb.DiscoveryMessage_ACTION_EXECUTE,
			Timestamp:      timestamppb.New(time.Now()),
			CurrentService: "client-service",
		}

		// Send request
		ctxg, cancel := ct.WithTimeout(ct.Background(), 5*time.Second)
		defer cancel()

		resg, err := c.ProcessDiscovery(ctxg, reqg)
		if err != nil {
			fmt.Printf("RPC failed: %v", err)
		}

		fmt.Printf("Response Status: %v", resg.Status)
		fmt.Printf("Response Service: %s", resg.CurrentService)
		fmt.Printf("Response Timestamp: %v", resg.Timestamp.AsTime())


*/
