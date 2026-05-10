package main

import (
	"context"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/llmutil"
	"github.com/charviki/maze/fabrication/cradle/logutil"

	"github.com/charviki/maze/the-mesa/the-forge/internal/config"
	"github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
	"github.com/charviki/maze/the-mesa/the-forge/internal/transport"
)

const defaultGRPCAddr = ":9090"

func main() {
	logger := logutil.New("the-forge")

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	// PostgreSQL 连接池
	pool, err := db.NewPool(context.Background(), db.PoolConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Name:     cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
	})
	if err != nil {
		logger.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	// Migration
	if err := db.RunMigrations(pool, postgres.MigrationsFS); err != nil {
		logger.Fatalf("run migrations: %v", err)
	}
	logger.Info("database migrations applied")

	// Repository
	knowledgeRepo := postgres.NewKnowledgeRepository(pool)
	taskRepo := postgres.NewTaskRepository(pool)
	chatRepo := postgres.NewChatRepository(pool)
	fileRepo := postgres.NewFileRepository(cfg.File.StoragePath)

	// Service
	knowledgeSvc := service.NewKnowledgeService(knowledgeRepo)
	taskSvc := service.NewTaskService(taskRepo)
	chatSvc := service.NewChatService(chatRepo, newLLMProvider(cfg))
	fileSvc := service.NewFileService(fileRepo)

	// gRPC server
	grpcCore := grpc.NewServer()
	knowledgeTransport := transport.NewKnowledgeGRPCTransport(knowledgeSvc, taskSvc)
	directiveTransport := transport.NewDirectiveGRPCTransport(taskSvc)
	knowledgeTransport.RegisterGRPC(grpcCore)
	directiveTransport.RegisterGRPC(grpcCore)
	managedGRPC := grpcutil.NewManagedGRPCServer(grpcAddr(cfg), grpcCore, logger)

	// grpc-gateway
	gwMux := gwruntime.NewServeMux()
	gwCtx, gwCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer gwCancel()
	if err := transport.RegisterGatewayHandlers(gwCtx, gwMux, grpcAddr(cfg)); err != nil {
		logger.Fatalf("register gateway: %v", err)
	}

	// HTTP server
	healthService := service.NewHealthService()
	httpServer := transport.NewHTTPServer(transport.HTTPServerParams{
		Config:        cfg,
		HealthService: healthService,
		ChatHandler:   transport.NewChatHandler(chatSvc, logger),
		FileHandler:   transport.NewFileHandler(fileSvc),
		GWMux:         gwMux,
		Logger:        logger,
	})

	// Lifecycle
	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpServer)
	manager.Add(managedGRPC)

	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run lifecycle: %v", err)
	}
}

func grpcAddr(cfg *config.Config) string {
	if cfg.Server.GRPCAddr == "" {
		return defaultGRPCAddr
	}
	return cfg.Server.GRPCAddr
}

func newLLMProvider(cfg *config.Config) llmutil.LLMProvider {
	providerCfg := llmutil.ProviderConfig{
		APIKey:  cfg.LLM.APIKey,
		BaseURL: cfg.LLM.BaseURL,
		Model:   cfg.LLM.Model,
	}
	switch cfg.LLM.Provider {
	case "anthropic":
		return llmutil.NewAnthropicProvider(providerCfg)
	case "openai":
		return llmutil.NewOpenAIProvider(providerCfg)
	default:
		// 默认使用 OpenAI 兼容接口
		return llmutil.NewOpenAIProvider(providerCfg)
	}
}
