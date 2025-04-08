package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/kubev2v/migration-planner/internal/agent"
	"github.com/kubev2v/migration-planner/internal/agent/config"
	"github.com/kubev2v/migration-planner/pkg/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	agentFilename = "agent_id"
	jwtFilename   = "jwt.json"
)

func main() {
	command := NewAgentCommand()
	if command == nil {
		os.Exit(1)
	}
	if err := command.Execute(); err != nil {
		zap.S().Fatalf("Agent execution failed: %v", err)
		os.Exit(1)
	}
}

type agentCmd struct {
	config     *config.Config
	configFile string
	logger     *zap.Logger
	logLevel   zap.AtomicLevel
}

func NewAgentCommand() *agentCmd {
	defaultLogLevel := zapcore.InfoLevel
	atomicLogLevel := zap.NewAtomicLevelAt(defaultLogLevel)

	logger := log.InitLog(atomicLogLevel)
	defer func() {
		_ = logger.Sync()
	}()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	a := &agentCmd{
		config:   config.NewDefault(),
		logger:   logger,
		logLevel: atomicLogLevel,
	}

	flag.StringVar(&a.configFile, "config", config.DefaultConfigFile, "Path to the agent's configuration file.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("This program starts an agent with the specified configuration. Below are the available flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if err := a.config.ParseConfigFile(a.configFile); err != nil {
		zap.S().Fatalf("Error parsing config file '%s': %v", a.configFile, err)
		return nil
	}
	if err := a.config.Validate(); err != nil {
		zap.S().Fatalf("Error validating config: %v", err)
		return nil
	}

	// --- TEMPORARY: Call simulation function ---
	// Simulate 100 logs with 100ms delay (10 logs/second)
	// Expect roughly 1 + (99 / 3) = 1 + 33 = 34 logs for "HTTP request completed"
	go simulateHighVolumeLogs(logger, 100, 100*time.Millisecond)
	zap.S().Infof("--- Finished high-volume log simulation ---")
	// Using a goroutine so it doesn't block the main startup flow entirely
	// Add a small sleep if needed to ensure some logs are generated before proceeding
	time.Sleep(1 * time.Second) // Give simulation a moment to start spitting logs

	return a
}

func (a *agentCmd) Execute() error {
	zap.S().Info("Executing agent command...")

	agentID, err := a.readFileFromPersistent(agentFilename)
	if err != nil {
		zap.S().Fatalf("failed to retreive agent_id: %v", err)
	}

	// Try to read jwt from file.
	// We're assuming the jwt is valid.
	// The agent will not try to validate the jwt. The backend is responsible for validating the token.
	jwt, err := a.readFileFromVolatile(jwtFilename)
	if err != nil {
		zap.S().Errorf("failed to read jwt: %v", err)
	}

	agentInstance := agent.New(uuid.MustParse(agentID), jwt, a.config)
	if err := agentInstance.Run(context.Background()); err != nil {
		zap.S().Fatalf("running device agent: %v", err)
	}
	return nil
}

func (a *agentCmd) readFile(baseDir string, filename string) (string, error) {
	filePath := path.Join(baseDir, filename)
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(bytes.TrimSpace(content)), nil
	}

	return "", fmt.Errorf("file not found: %s", filePath)
}

func (a *agentCmd) readFileFromVolatile(filename string) (string, error) {
	return a.readFile(a.config.DataDir, filename)
}

func (a *agentCmd) readFileFromPersistent(filename string) (string, error) {
	return a.readFile(a.config.PersistentDataDir, filename)
}

func simulateHighVolumeLogs(logger *zap.Logger, count int, delay time.Duration) {
	zap.S().Infof("--- Starting high-volume log simulation (%d logs, %v delay) ---", count, delay)

	staticFields := []zap.Field{
		zap.String("type", "http_request"),
		zap.String("request_id", "SIMULATED"),
		zap.String("http_method", "PUT"),
		zap.String("http_path", "/api/v1/agents/SIMULATED/status"),
		zap.String("http_proto", "HTTP/1.1"),
		zap.String("remote_addr", "[::1]:99999"),
		zap.Int("http_status_code", 200),
		zap.String("http_status_text", "200 OK"),
		zap.Int64("response_bytes", 0),
		zap.String("user_agent", "SimulationClient/1.0"),
	}

	for i := 0; i < count; i++ {
		simulatedLatency := time.Duration(rand.Intn(5)+1) * time.Millisecond // 1-5 ms
		dynamicFields := append(staticFields, zap.Duration("latency", simulatedLatency))

		var httpPath string
		for _, field := range staticFields {
			if field.Key == "http_path" {
				httpPath = field.String
				break
			}
		}

		msg := fmt.Sprintf("******HTTP request completed for path: %s******", httpPath)
		logger.Info(msg, dynamicFields...)

		time.Sleep(delay)
	}

	zap.S().Infof("--- Finished high-volume log simulation ---")
	_ = logger.Sync()
}
