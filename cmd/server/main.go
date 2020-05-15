package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/allisson/go-env"
	"github.com/allisson/go-pglock"
	pb "github.com/allisson/hammer/api/v1"
	hammerGrpc "github.com/allisson/hammer/grpc"
	repository "github.com/allisson/hammer/repository/postgres"
	"github.com/allisson/hammer/service"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger       *zap.Logger
	sqlDB        *sqlx.DB
	grpcEndpoint string
	httpEndpoint string
)

func init() {
	// Set logger
	logger, _ = zap.NewProduction()

	// Set database connection
	db, err := sqlx.Open("postgres", env.GetString("HAMMER_DATABASE_URL", ""))
	if err != nil {
		logger.Fatal("failed-to-start-database-client", zap.Error(err))
	}
	err = db.Ping()
	if err != nil {
		logger.Fatal("failed-to-ping-database", zap.Error(err))
	}
	sqlDB = db

	// Set grpc endpoint
	grpcEndpoint = fmt.Sprintf(":%d", env.GetInt("HAMMER_GRPC_PORT", 50051))

	// Set http endpoint
	httpEndpoint = fmt.Sprintf(":%d", env.GetInt("HAMMER_HTTP_PORT", 8000))
}

func runGateway() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterHammerHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		logger.Error("http-serve", zap.Error(err))
		return
	}

	if err := http.ListenAndServe(httpEndpoint, mux); err != nil {
		logger.Error("http-serve", zap.Error(err))
	}
}

func metricsServer() {
	port := env.GetInt("HAMMER_METRICS_PORT", 4001)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		logger.Error("metrics-server-failed-to-start", zap.Error(err))
	}
}

func databaseMigrate() error {
	// Get conn
	conn, err := sqlDB.DB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	// Is the answer to everything
	lockID := int64(42)

	// Create a new lock
	lock := pglock.NewLock(conn)

	// Lock the resource
	ok, err := lock.Lock(lockID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	// Unlock the resource
	defer func() {
		if err := lock.Unlock(lockID); err != nil {
			logger.Error("database-migrations-unlock", zap.Error(err))
		}
	}()

	// Execute migration
	driver, err := postgres.WithInstance(sqlDB.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(env.GetString("HAMMER_DATABASE_MIGRATION_DIR", "file:///db/migrations"), "postgres", driver)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			return err
		}
	}
	logger.Info("database-migrations-completed")
	return nil
}

func main() {
	// Create repositories
	topicRepo := repository.NewTopic(sqlDB)
	subscriptionRepo := repository.NewSubscription(sqlDB)
	messageRepo := repository.NewMessage(sqlDB)
	deliveryRepo := repository.NewDelivery(sqlDB)
	deliveryAttemptRepo := repository.NewDeliveryAttempt(sqlDB)
	txFactoryRepo := repository.NewTxFactory(sqlDB)

	// Create services
	topicService := service.NewTopic(&topicRepo, &txFactoryRepo)
	subscriptionService := service.NewSubscription(&topicRepo, &subscriptionRepo, &txFactoryRepo)
	messageService := service.NewMessage(&topicRepo, &messageRepo, &subscriptionRepo, &deliveryRepo, &txFactoryRepo)
	deliveryService := service.NewDelivery(&deliveryRepo, &deliveryAttemptRepo, &txFactoryRepo)
	deliveryAttemptService := service.NewDeliveryAttempt(&deliveryAttemptRepo)

	// Create grpc handlers
	topicHandler := hammerGrpc.NewTopicHandler(&topicService)
	subscriptionHandler := hammerGrpc.NewSubscriptionHandler(&subscriptionService)
	messageHandler := hammerGrpc.NewMessageHandler(&messageService)
	deliveryHandler := hammerGrpc.NewDeliveryHandler(&deliveryService)
	deliveryAttemptHandler := hammerGrpc.NewDeliveryAttemptHandler(&deliveryAttemptService)

	// Execute database migrations
	err := databaseMigrate()
	if err != nil {
		logger.Fatal("database-migrations-error", zap.Error(err))
	}

	// Start tcp server
	listener, err := net.Listen("tcp", grpcEndpoint)
	if err != nil {
		logger.Fatal("failed-to-listen", zap.Error(err))
	}

	// Start http gateway
	go runGateway()

	// Create grpc server
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	server := hammerGrpc.NewServer(topicHandler, subscriptionHandler, messageHandler, deliveryHandler, deliveryAttemptHandler)

	// Register grpc services
	pb.RegisterHammerServer(grpcServer, &server)

	// Enable grpc_prometheus histograms
	grpc_prometheus.EnableHandlingTimeHistogram()

	// Register grpc_prometheus default metrics
	grpc_prometheus.Register(grpcServer)

	// Load metrics server
	go metricsServer()

	// Start grpc server and make graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		// We received an interrupt signal, shut down.
		logger.Info("grpc-server-shutdown-started")
		grpcServer.GracefulStop()
		close(idleConnsClosed)
		logger.Info("grpc-server-shutdown-finished")
	}()

	if err := grpcServer.Serve(listener); err != nil {
		logger.Error("grpc-server-serve", zap.Error(err))
	}

	<-idleConnsClosed
}
