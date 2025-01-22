# ModuloX

ModuloX is a modular AI framework built in Go that provides a flexible and extensible system for building AI agents and multi-agent systems.

## Features

- Modular architecture with clean interfaces
- Support for multiple LLM providers
- Vector storage for agent memory
- Tool integration system
- Multi-agent orchestration
- Built-in reliability features

## Core Components

- Agent Interface: Defines basic agent behaviors including task execution, tool addition, and capability discovery
- LLM Provider Interface: Supports integration with multiple language models
- Vector Store Interface: Manages agent memory and knowledge
- Tool Registry: Handles tool registration and discovery
- Workflow Engine: Supports sequential and parallel execution
- Communication System: Enables agent collaboration

## Getting Started

```bash
# Clone the repository
git clone https://github.com/sanfactor/shengyenatie.git

# Build the project
cd shengyenatie
go build ./cmd/modulox

# Run the agent
./modulox
```

## Documentation

See the `docs` directory for detailed documentation on:
- Architecture Overview
- Component Interfaces
- Configuration Guide
- Tool Development
- Multi-Agent Patterns

## License

MIT License
