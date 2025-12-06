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
	"github.com/abhissng/neuron/adapters/paseto"
	neuronctx "github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
	"github.com/abhissng/neuron/utils/structures/claims"
	"github.com/abhissng/neuron/utils/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServiceRegistrar is a callback function for registering gRPC services.
// It receives the underlying grpc.Server so users can register their proto services.
type ServiceRegistrar func(server *grpc.Server)

// CustomValidatorFunc is a function signature for external validation logic.
// It is called after token parsing but before the handler.
// Parameters:
//   - ctx: the context with claims already populated
//   - fullMethod: the full gRPC method name (e.g., "/package.Service/Method")
//   - claims: the parsed claims from the token
//
// Return an error to reject the request.
type CustomValidatorFunc func(ctx context.Context, fullMethod string, claims *claims.StandardClaims) error

// Server represents a gRPC server
type Server struct {
	server *grpc.Server
	config ServerConfig
}

// NeuronServer is an enhanced gRPC server wrapper with lifecycle management,
// ServiceContext propagation, Paseto authentication, and external validation hooks.
type NeuronServer struct {
	*Server
	listener net.Listener
}

// NewNeuronServer creates a new NeuronServer with the provided options.
// It wraps the standard Server with additional lifecycle management.
func NewNeuronServer(opts ...Option) (*NeuronServer, error) {
	s, err := NewServer(opts...)
	if err != nil {
		return nil, err
	}

	// Register services if a registrar was provided
	if s.config.serviceRegistrar != nil {
		s.config.serviceRegistrar(s.server)
	}

	return &NeuronServer{Server: s}, nil
}

// Start starts the NeuronServer and blocks until stopped.
func (ns *NeuronServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", ns.config.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	ns.listener = lis
	ns.config.log.Info(fmt.Sprintf("NeuronServer starting on port %d", ns.config.port),
		zap.String("service", ns.config.serviceName),
		zap.String("auth_mode", ns.config.authMode),
	)
	return ns.server.Serve(lis)
}

// Stop gracefully stops the NeuronServer.
func (ns *NeuronServer) Stop() {
	ns.GracefulStop()
}

// GetGRPCServer returns the underlying grpc.Server for advanced use cases.
func (ns *NeuronServer) GetGRPCServer() *grpc.Server {
	return ns.server
}

// RegisterService allows registering additional services after creation.
func (ns *NeuronServer) RegisterService(registrar ServiceRegistrar) {
	registrar(ns.server)
}

// NewServer creates a new gRPC server with option pattern
func NewServer(opts ...Option) (*Server, error) {
	// Default configuration
	config := ServerConfig{
		port:           50051,
		serviceName:    "default-service",
		maxRecvMsgSize: 4, // Default 4MB
		maxSendMsgSize: 4, // Default 4MB
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
		grpc.MaxSendMsgSize(config.maxSendMsgSize*1024*1024),
	)

	// Create gRPC Server
	s := &Server{
		server: grpc.NewServer(grpcOpts...),
		config: config,
	}

	if config.enableMetrics {
		grpc_prometheus.Register(s.server)
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

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(recoveryHandler),
	}
	unary = append(unary, recovery.UnaryServerInterceptor(recoveryOpts...))
	stream = append(stream, recovery.StreamServerInterceptor(recoveryOpts...))

	unary = append(unary, unaryCorrelationIDInterceptor())
	stream = append(stream, streamCorrelationIDInterceptor())

	unary = append(unary, unaryRequestIDInterceptor())
	stream = append(stream, streamRequestIDInterceptor())

	// Add ServiceContext propagation interceptor
	if config.appContext != nil {
		unary = append(unary, unaryServiceContextInterceptor(config.appContext))
		stream = append(stream, streamServiceContextInterceptor(config.appContext))
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	unary = append(unary, logging.UnaryServerInterceptor(InterceptorLogger(config.log), loggingOpts...))
	stream = append(stream, logging.StreamServerInterceptor(InterceptorLogger(config.log), loggingOpts...))

	switch config.authMode {
	case "jwt":
		if config.jwtSecret != "" {
			authFunc := createAuthFunc(config.jwtSecret)
			unary = append(unary, auth.UnaryServerInterceptor(authFunc))
			stream = append(stream, auth.StreamServerInterceptor(authFunc))
		}
	case "paseto":
		if config.pasetoManager != nil {
			// Use enhanced Paseto auth with custom validator support
			unary = append(unary, unaryPasetoAuthInterceptor(config))
			stream = append(stream, streamPasetoAuthInterceptor(config))
		}
	}

	if config.enableMetrics {
		grpc_prometheus.EnableHandlingTimeHistogram()
		unary = append(unary, grpc_prometheus.UnaryServerInterceptor)
		stream = append(stream, grpc_prometheus.StreamServerInterceptor)
	}

	return unary, stream
}

func unaryCorrelationIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := ctx.Value(constant.CorrelationID).(types.StringConstant); !ok {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if vals := md.Get(constant.CorrelationID); len(vals) > 0 {
					ctx = context.WithValue(ctx, types.StringConstant(constant.CorrelationID), vals[0])
				}
			}
		}
		if _, ok := ctx.Value(types.StringConstant(constant.CorrelationID)).(string); !ok {
			ctx = context.WithValue(ctx, types.StringConstant(constant.CorrelationID), random.GenerateUUID())
		}
		return handler(ctx, req)
	}
}

func streamCorrelationIDInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if _, ok := ss.Context().Value(constant.CorrelationID).(types.StringConstant); !ok {
			if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
				if vals := md.Get(constant.CorrelationID); len(vals) > 0 {
					newCtx := context.WithValue(ss.Context(), types.StringConstant(constant.CorrelationID), vals[0])
					wrapped := &serverStreamWithContext{ServerStream: ss, ctx: newCtx}
					return handler(srv, wrapped)
				}
			}
			// metadata missing value, generate correlation id
			newCtx := context.WithValue(ss.Context(), types.StringConstant(constant.CorrelationID), random.GenerateUUID())
			wrapped := &serverStreamWithContext{ServerStream: ss, ctx: newCtx}
			return handler(srv, wrapped)
		}
		if _, ok := ss.Context().Value(types.StringConstant(constant.CorrelationID)).(string); !ok {
			newCtx := context.WithValue(ss.Context(), types.StringConstant(constant.CorrelationID), random.GenerateUUID())
			wrapped := &serverStreamWithContext{ServerStream: ss, ctx: newCtx}
			return handler(srv, wrapped)
		}
		return handler(srv, ss)
	}
}

func unaryRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := ctx.Value(constant.RequestID).(types.StringConstant); !ok {
			ctx = context.WithValue(ctx, types.StringConstant(constant.RequestID), random.GenerateUUID())
		}
		return handler(ctx, req)
	}
}

func streamRequestIDInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if _, ok := ss.Context().Value(constant.RequestID).(types.StringConstant); !ok {
			ctx := context.WithValue(ss.Context(), types.StringConstant(constant.RequestID), random.GenerateUUID())
			wrapped := &serverStreamWithContext{ServerStream: ss, ctx: ctx}
			return handler(srv, wrapped)
		}
		return handler(srv, ss)
	}
}

type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serverStreamWithContext) Context() context.Context { return w.ctx }

// InterceptorLogger is a simple logging manager
func InterceptorLogger(l *log.Log) logging.Logger {
	// convert key-value pairs from interceptor to zap fields
	toZapFields := func(kvs ...any) []zap.Field {
		zfs := make([]zap.Field, 0, len(kvs)/2)
		for i := 0; i+1 < len(kvs); i += 2 {
			key, ok := kvs[i].(string)
			if !ok {
				key = fmt.Sprintf("arg_%d", i)
			}
			zfs = append(zfs, zap.Any(key, kvs[i+1]))
		}
		return zfs
	}

	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		requestID, _ := ctx.Value(constant.RequestID).(types.StringConstant)
		correlationID, _ := ctx.Value(constant.CorrelationID).(types.StringConstant)

		zFields := []zap.Field{
			zap.String("correlation_id", string(correlationID)),
			zap.String("request_id", string(requestID)),
		}
		zFields = append(zFields, toZapFields(fields...)...)

		// #nosec: G115 - We are intentionally converting logging.Level to zapcore.Level
		// The logging.Level is guaranteed to be within the valid range for zapcore.Level
		// This is safe because logging.Level values are constrained to valid zapcore.Level values
		switch zapcore.Level(lvl) {
		case zapcore.DebugLevel:
			l.Debug(msg, zFields...)
		case zapcore.InfoLevel:
			l.Info(msg, zFields...)
		case zapcore.WarnLevel:
			l.Warn(msg, zFields...)
		case zapcore.ErrorLevel:
			l.Error(msg, zFields...)
		case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
			l.Error(msg, zFields...)
		default:
			l.Info(msg, zFields...)
		}
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

// unaryServiceContextInterceptor creates a ServiceContext and stores it in the context.
func unaryServiceContextInterceptor(appCtx *neuronctx.AppContext) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Create ServiceContext with AppContext
		svcCtx := neuronctx.NewServiceContext(
			neuronctx.WithAppContext(appCtx),
		)

		// Propagate request ID and correlation ID from gRPC context
		if reqID, ok := ctx.Value(types.StringConstant(constant.RequestID)).(string); ok {
			svcCtx = svcCtx.WithRequestID(reqID)
		}
		if corrID, ok := ctx.Value(types.StringConstant(constant.CorrelationID)).(string); ok {
			svcCtx = svcCtx.WithValue(constant.CorrelationID, corrID)
		}

		// Store ServiceContext in the context
		ctx = context.WithValue(ctx, types.StringConstant(constant.ServiceContext), svcCtx)

		return handler(ctx, req)
	}
}

// streamServiceContextInterceptor creates a ServiceContext for stream handlers.
func streamServiceContextInterceptor(appCtx *neuronctx.AppContext) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Create ServiceContext with AppContext
		svcCtx := neuronctx.NewServiceContext(
			neuronctx.WithAppContext(appCtx),
		)

		// Propagate request ID and correlation ID from gRPC context
		if reqID, ok := ctx.Value(types.StringConstant(constant.RequestID)).(string); ok {
			svcCtx = svcCtx.WithRequestID(reqID)
		}
		if corrID, ok := ctx.Value(types.StringConstant(constant.CorrelationID)).(string); ok {
			svcCtx = svcCtx.WithValue(constant.CorrelationID, corrID)
		}

		// Store ServiceContext in the context
		newCtx := context.WithValue(ctx, types.StringConstant(constant.ServiceContext), svcCtx)
		wrapped := &serverStreamWithContext{ServerStream: ss, ctx: newCtx}

		return handler(srv, wrapped)
	}
}

// unaryPasetoAuthInterceptor handles Paseto authentication with custom validator support.
func unaryPasetoAuthInterceptor(config ServerConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if method should skip auth
		if config.skipAuthMethods != nil && config.skipAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract and validate token
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			config.log.Warn("Missing or invalid authorization header",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid authorization")
		}

		res := config.pasetoManager.ValidateToken(token, nil, paseto.WithValidateEssentialTags)
		if res.IsFailure() {
			config.log.Warn("Invalid Paseto token",
				zap.String("method", info.FullMethod),
			)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		cl, err := res.Value()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to extract claims")
		}

		// Populate context with claims
		ctx = populateContextWithClaims(ctx, cl)

		// Store claims in context for custom validator
		ctx = context.WithValue(ctx, types.StringConstant(constant.Claims), cl)

		// Call custom validator if provided
		if config.customValidator != nil {
			if err := config.customValidator(ctx, info.FullMethod, cl); err != nil {
				config.log.Warn("Custom validation failed",
					zap.String("method", info.FullMethod),
					zap.String("user_id", cl.Sub),
					zap.Error(err),
				)
				return nil, status.Errorf(codes.PermissionDenied, "validation failed: %v", err)
			}
		}

		config.log.Debug("Request authenticated",
			zap.String("method", info.FullMethod),
			zap.String("user_id", cl.Sub),
		)

		return handler(ctx, req)
	}
}

// streamPasetoAuthInterceptor handles Paseto authentication for streams.
func streamPasetoAuthInterceptor(config ServerConfig) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Check if method should skip auth
		if config.skipAuthMethods != nil && config.skipAuthMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// Extract and validate token
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			config.log.Warn("Missing or invalid authorization header",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return status.Errorf(codes.Unauthenticated, "missing or invalid authorization")
		}

		res := config.pasetoManager.ValidateToken(token, nil, paseto.WithValidateEssentialTags)
		if res.IsFailure() {
			config.log.Warn("Invalid Paseto token",
				zap.String("method", info.FullMethod),
			)
			return status.Errorf(codes.Unauthenticated, "invalid token")
		}

		cl, err := res.Value()
		if err != nil {
			return status.Errorf(codes.Internal, "failed to extract claims")
		}

		// Populate context with claims
		ctx = populateContextWithClaims(ctx, cl)

		// Store claims in context for custom validator
		ctx = context.WithValue(ctx, types.StringConstant(constant.Claims), cl)

		// Call custom validator if provided
		if config.customValidator != nil {
			if err := config.customValidator(ctx, info.FullMethod, cl); err != nil {
				config.log.Warn("Custom validation failed",
					zap.String("method", info.FullMethod),
					zap.String("user_id", cl.Sub),
					zap.Error(err),
				)
				return status.Errorf(codes.PermissionDenied, "validation failed: %v", err)
			}
		}

		config.log.Debug("Stream authenticated",
			zap.String("method", info.FullMethod),
			zap.String("user_id", cl.Sub),
		)

		wrapped := &serverStreamWithContext{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

// populateContextWithClaims adds claim values to the context.
func populateContextWithClaims(ctx context.Context, cl *claims.StandardClaims) context.Context {
	if cl == nil {
		return ctx
	}

	if cl.Sub != "" {
		ctx = context.WithValue(ctx, types.StringConstant(constant.UserID), cl.Sub)
	}
	if cl.Iss != "" {
		ctx = context.WithValue(ctx, types.StringConstant(constant.Issuer), cl.Iss)
	}
	if cl.Jti != "" {
		ctx = context.WithValue(ctx, types.StringConstant(constant.TokenID), cl.Jti)
	}
	if cl.Data != nil {
		// if svc, ok := cl.Data["service"].(string); ok {
		// 	ctx = context.WithValue(ctx, types.StringConstant(constant.Service), svc)
		// }
		// if roles, ok := cl.Data["roles"].([]string); ok {
		// 	ctx = context.WithValue(ctx, types.StringConstant(constant.Roles), roles)
		// }
		// Store all custom data
		ctx = context.WithValue(ctx, types.StringConstant(constant.ClaimsData), cl.Data)
	}

	// Ensure request ID is set
	if _, ok := ctx.Value(types.StringConstant(constant.RequestID)).(string); !ok {
		ctx = context.WithValue(ctx, types.StringConstant(constant.RequestID), random.GenerateUUID())
	}

	return ctx
}

// GetServiceContextFromContext extracts the ServiceContext from a gRPC context.
// Returns nil if not found.
func GetServiceContextFromContext(ctx context.Context) *neuronctx.ServiceContext {
	if svcCtx, ok := ctx.Value(types.StringConstant(constant.ServiceContext)).(*neuronctx.ServiceContext); ok {
		return svcCtx
	}
	return nil
}

// GetClaimsFromContext extracts the StandardClaims from a gRPC context.
// Returns nil if not found.
func GetClaimsFromContext(ctx context.Context) *claims.StandardClaims {
	if cl, ok := ctx.Value(types.StringConstant(constant.Claims)).(*claims.StandardClaims); ok {
		return cl
	}
	return nil
}

// GetUserIDFromContext extracts the user ID from a gRPC context.
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(types.StringConstant(constant.UserID)).(string); ok {
		return userID
	}
	return ""
}

// GetRolesFromContext extracts the roles from a gRPC context.
func GetRolesFromContext(ctx context.Context) []string {
	if roles, ok := ctx.Value(types.StringConstant(constant.Roles)).([]string); ok {
		return roles
	}
	return nil
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

/*
USAGE EXAMPLES

// ===== Server: Enable JWT auth and log correlation_id =====
func startJWTServer() error {
    s, err := NewServer(
        WithPort(50051),
        WithAuthMode("jwt"),
        WithJWT("<your-jwt-secret>"),
        WithMetrics(),
    )
    if err != nil { return err }

    // register your protobuf service servers here, e.g.:
    // pb.RegisterYourServiceServer(s.server, yourImpl)

    return s.Start()
}

// ===== Server: Enable PASETO auth =====
func startPasetoServer(pubKey ed25519.PublicKey) error {
    pm := paseto.NewPasetoManager(
        paseto.WithPublicKey(pubKey),
        paseto.WithIssuer("your-issuer"),
        paseto.WithAccessTokenExpiry(time.Hour),
    )
    s, err := NewServer(
        WithPort(50051),
        WithAuthMode("paseto"),
        WithPasetoManager(pm),
    )
    if err != nil { return err }
    // pb.RegisterYourServiceServer(s.server, yourImpl)
    return s.Start()
}

// ===== Unary handler: read correlation_id/request_id from context =====
func (h *handler) SomeRPC(ctx context.Context, req *pb.SomeRequest) (*pb.SomeReply, error) {
    corr, _ := ctx.Value(types.StringConstant(constant.CorrelationID)).(types.StringConstant)
    reqID, _ := ctx.Value(types.StringConstant(constant.RequestID)).(types.StringConstant)
    h.log.Info("handling request", "correlation_id", corr, "request_id", reqID)
    // ...
    return &pb.SomeReply{}, nil
}

// ===== Client: attach Authorization and correlation_id =====
func callWithJWT(conn *grpc.ClientConn, token string) error {
    c := pb.NewYourServiceClient(conn)
    md := metadata.New(map[string]string{
        "authorization":  "Bearer " + token,
        "correlation_id": "corr-1234",
    })
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    _, err := c.SomeRPC(ctx, &pb.SomeRequest{})
    return err
}
*/
