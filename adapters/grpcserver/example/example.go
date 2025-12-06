// Package main demonstrates how to use the NeuronServer gRPC wrapper.
// This example shows:
// - Creating a NeuronServer with Paseto authentication
// - Injecting custom validation logic from outside the library
// - Using ServiceContext propagation
// - Registering gRPC services via callback
package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcmanager "github.com/abhissng/neuron/adapters/grpcserver"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/paseto"
	neuronctx "github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/structures/claims"
	"google.golang.org/grpc"
	// Import your proto-generated packages here:
	// pb "your-module/proto/yourservice"
)

// =============================================================================
// EXAMPLE: Custom Validator Function
// =============================================================================

// myCustomValidator is an example of external validation logic that can be
// injected into the NeuronServer without modifying the library code.
// This function is called AFTER token parsing but BEFORE the handler.
func myCustomValidator(ctx context.Context, fullMethod string, cl *claims.StandardClaims) error {
	// Example 1: Block specific methods for certain users
	if fullMethod == "/mypackage.AdminService/DeleteUser" {
		roles := grpcmanager.GetRolesFromContext(ctx)
		if !containsRole(roles, "admin") {
			return errors.New("admin role required for this operation")
		}
	}

	// Example 2: Check if user is active in database
	// userID := cl.Sub
	// if !isUserActive(userID) {
	//     return errors.New("user account is suspended")
	// }

	// Example 3: Rate limiting based on user
	// if isRateLimited(cl.Sub, fullMethod) {
	//     return errors.New("rate limit exceeded")
	// }

	// Example 4: Method-specific validation
	switch fullMethod {
	case "/mypackage.PaymentService/ProcessPayment":
		// Require additional verification for payment methods
		if cl.Data == nil || cl.Data["payment_verified"] != true {
			return errors.New("payment verification required")
		}
	}

	return nil // Validation passed
}

func containsRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

// =============================================================================
// EXAMPLE: Service Registration Callback
// =============================================================================

// registerServices is a callback that registers your gRPC services.
// This decouples proto implementations from the NeuronServer library.
func registerServices(server *grpc.Server) {
	// Register your protobuf service implementations here:
	// pb.RegisterYourServiceServer(server, &yourServiceImpl{})
	// pb.RegisterAnotherServiceServer(server, &anotherServiceImpl{})

	fmt.Println("Services registered successfully")
}

// =============================================================================
// EXAMPLE: gRPC Handler using ServiceContext
// =============================================================================

// ExampleHandler shows how to use ServiceContext in your gRPC handlers.
// type myServiceServer struct {
// 	pb.UnimplementedMyServiceServer
// }
//
// func (s *myServiceServer) SomeMethod(ctx context.Context, req *pb.SomeRequest) (*pb.SomeResponse, error) {
// 	// Get the ServiceContext from the gRPC context
// 	svcCtx := grpcmanager.GetServiceContextFromContext(ctx)
// 	if svcCtx != nil {
// 		// Now you have access to all AppContext resources:
// 		// - svcCtx.Log (logger)
// 		// - svcCtx.Database
// 		// - svcCtx.RedisManager
// 		// - svcCtx.PasetoManager
// 		// etc.
//
// 		svcCtx.SlogInfo("Processing request", log.String("user_id", grpcmanager.GetUserIDFromContext(ctx)))
// 	}
//
// 	// Get claims directly if needed
// 	claims := grpcmanager.GetClaimsFromContext(ctx)
// 	if claims != nil {
// 		fmt.Printf("Request from user: %s\n", claims.Sub)
// 	}
//
// 	return &pb.SomeResponse{}, nil
// }

// =============================================================================
// MAIN: Server Setup and Lifecycle
// =============================================================================

func main() {
	// 1. Generate or load your Ed25519 keys for Paseto
	// In production, load these from secure storage (Vault, env vars, etc.)
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		fmt.Printf("Failed to generate keys: %v\n", err)
		os.Exit(1)
	}

	// 2. Create the Paseto manager
	pasetoManager := paseto.NewPasetoManager(
		paseto.WithKeys(privateKey, publicKey),
		paseto.WithIssuer("my-service"),
		paseto.WithAccessTokenExpiry(1*time.Hour),
		paseto.WithRefreshTokenExpiry(24*time.Hour),
	)

	// 3. Create the logger
	logger, err := log.NewLogger(log.NewLoggerConfig(false,
		log.WithServiceName("my-grpc-service"),
		log.WithDisableOpenSearch(true),
	))
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// 4. Create the AppContext with all your dependencies
	appCtx := neuronctx.NewAppContext(
		neuronctx.WithLogger(logger),
		neuronctx.WithServiceID("my-grpc-service"),
		// Add other dependencies as needed:
		// neuronctx.WithDatabase(db),
		// neuronctx.WithRedisManager(redis),
	)

	// 5. Create the NeuronServer with all options
	server, err := grpcmanager.NewNeuronServer(
		// Basic configuration
		grpcmanager.WithPort(50051),
		grpcmanager.WithServiceName("my-grpc-service"),
		grpcmanager.WithLogger(logger),

		// Authentication
		grpcmanager.WithAuthMode("paseto"),
		grpcmanager.WithPasetoManager(pasetoManager),

		// Skip auth for health checks and public methods
		grpcmanager.WithSkipAuthMethods(
			"/grpc.health.v1.Health/Check",
			"/mypackage.PublicService/GetPublicData",
		),

		// Context propagation - makes ServiceContext available in handlers
		grpcmanager.WithAppContext(appCtx),

		// External validation logic - called after token parsing
		grpcmanager.WithCustomValidator(myCustomValidator),

		// Service registration callback - decouples proto from library
		grpcmanager.WithServiceRegistrar(registerServices),

		// Optional: Enable Prometheus metrics
		grpcmanager.WithMetrics(),

		// Optional: TLS configuration
		// grpcmanager.WithTLS("cert.pem", "key.pem", "ca.pem"),
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create server: %v", err))
	}

	// 6. Start the server in a goroutine
	go func() {
		logger.Info("Starting NeuronServer...")
		if err := server.Start(); err != nil {
			logger.Fatal(fmt.Sprintf("Server failed: %v", err))
		}
	}()

	// 7. Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	server.Stop()
	logger.Info("Server stopped")
}

// =============================================================================
// EXAMPLE: Generating a token for testing
// =============================================================================

// generateTestToken shows how to generate a Paseto token for testing.
// In production, this would be done by your auth service.
/*
func generateTestToken(pm *paseto.PasetoManager) (string, error) {
	result := pm.FetchToken(
		claims.WithSubject("user-123"),
		claims.WithData(map[string]any{
			"service": "my-service",
			"roles":   []string{"user", "admin"},
		}),
	)

	if result.IsFailure() {
		return "", errors.New("failed to generate token")
	}

	tokenDetails, err := result.Value()
	if err != nil {
		return "", err
	}

	return tokenDetails.Token, nil
}
*/

// =============================================================================
// EXAMPLE: Client-side usage
// =============================================================================

// Example of how a client would call the server with authentication:
//
// func callServer() error {
//     conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
//     if err != nil {
//         return err
//     }
//     defer conn.Close()
//
//     client := pb.NewMyServiceClient(conn)
//
//     // Create context with auth token and correlation ID
//     md := metadata.New(map[string]string{
//         "authorization":  "Bearer " + token,
//         "correlation_id": "corr-12345",
//     })
//     ctx := metadata.NewOutgoingContext(context.Background(), md)
//
//     // Make the call
//     resp, err := client.SomeMethod(ctx, &pb.SomeRequest{})
//     return err
// }
