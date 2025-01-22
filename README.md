# ModuloX Framework

<div align="center">
<img src="docs/images/modulox-logo.png" alt="ModuloX Logo" width="300"/>

*The Enterprise-Grade Production-Ready Multi-Agent Framework in Go*

[![Go Version](https://img.shields.io/github/go-mod/go-version/sanfactor/shengyenatie)](https://github.com/sanfactor/shengyenatie)
[![License](https://img.shields.io/github/license/sanfactor/shengyenatie)](https://github.com/sanfactor/shengyenatie/blob/main/LICENSE)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](https://github.com/sanfactor/shengyenatie/tree/main/docs)
</div>

## ✨ Features

| Category | Features | Benefits |
|----------|----------|----------|
| 🏢 Enterprise Architecture | • Production-Ready Infrastructure<br>• High Reliability Systems<br>• Modular Design<br>• Comprehensive Observability | • Reduced downtime<br>• Easier maintenance<br>• Better monitoring<br>• Enhanced debugging |
| 🤖 Agent System | • LLM Integration<br>• Memory Management<br>• Tool Registry<br>• Plugin System | • Flexible AI models<br>• Long-term memory<br>• Extensible capabilities<br>• Custom tools |
| 🔄 Workflow Engine | • Sequential Processing<br>• Parallel Execution<br>• Dynamic Scheduling<br>• Result Aggregation | • Complex task handling<br>• Improved performance<br>• Flexible workflows<br>• Optimized execution |
| 📊 Observability | • Structured Logging<br>• Distributed Tracing<br>• Metrics Collection<br>• Health Checks | • Better debugging<br>• Performance insights<br>• System monitoring<br>• Proactive maintenance |
| 🛡️ Reliability | • Circuit Breaker<br>• Rate Limiting<br>• Automatic Retries<br>• Error Recovery | • Failure isolation<br>• System protection<br>• Enhanced stability<br>• Graceful degradation |
| 🌐 Communication | • gRPC Integration<br>• Event System<br>• Message Bus<br>• State Management | • Efficient communication<br>• Event-driven architecture<br>• Reliable messaging<br>• Distributed state |

## Requirements

- Go 1.18 or higher
- Protocol Buffers compiler (for gRPC)
- Access to LLM providers (optional)

## Installation

```bash
go get github.com/sanfactor/shengyenatie
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/sanfactor/shengyenatie/pkg/agent"
    "github.com/sanfactor/shengyenatie/pkg/llm"
    "github.com/sanfactor/shengyenatie/pkg/memory"
    "github.com/sanfactor/shengyenatie/pkg/tools"
)

func main() {
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

    // Execute task
    result, err := agent.Execute(context.Background(), "Analyze this data...")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result)
}
```

## Core Components

### Agent System
The agent system provides a flexible framework for creating AI agents with:
- LLM integration for natural language processing
- Vector store-based memory for context retention
- Type-safe tool registry for extending capabilities
- Plugin system for custom functionality

### Workflow Engine
Orchestrate multiple agents with:
- Sequential workflows for dependent tasks
- Parallel execution for independent operations
- Dynamic task scheduling and load balancing
- Result aggregation and processing

### Observability
Comprehensive monitoring with:
- Structured logging with context tracking
- Distributed tracing for request flows
- Metrics collection and visualization
- Health check system with automated recovery

### Reliability
Enterprise-grade reliability features:
- Circuit breaker pattern for failure isolation
- Token bucket rate limiting
- Exponential backoff retries
- Graceful error handling and recovery

### Communication
Advanced communication infrastructure:
- gRPC-based service communication
- Event-driven architecture
- Message bus for reliable messaging
- Distributed state management

## Documentation

- [Getting Started](docs/getting-started.md)
- [Architecture Overview](docs/architecture.md)
- [API Reference](docs/api/README.md)
- [Examples](examples/README.md)

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

This project is licensed under the [MIT License](LICENSE).
