package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/modulox/pkg/agent"
	"github.com/user/modulox/pkg/llm"
	"github.com/user/modulox/pkg/memory"
	"github.com/user/modulox/pkg/tools"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Initialize components
	provider := &llm.BaseProvider{
		Config: llm.ProviderConfig{
			ModelName: "gpt-3.5-turbo",
			MaxTokens: 4096,
		},
	}

	store := memory.NewBaseStore()
	registry := tools.NewToolRegistry()

	// Create base agent
	agent := agent.NewBaseAgent(agent.BaseAgentConfig{
		Name:        "modulox-agent",
		Description: "ModuloX framework base agent",
		Provider:    provider,
		Memory:      store,
		Registry:    registry,
	})

	// Run agent until context is cancelled
	fmt.Println("ModuloX agent started. Press Ctrl+C to exit.")
	<-ctx.Done()
	fmt.Println("\nShutting down...")
}
