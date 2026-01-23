// Package main provides the entry point for the SAM bridge server.
// The SAM bridge implements the SAMv3.3 protocol, enabling applications
// to communicate over the I2P anonymity network using text-based commands.
//
// Usage:
//
//	sam-bridge [flags]
//
// Flags:
//
//	-listen string     SAM listen address (default ":7656")
//	-i2cp string       I2CP router address (default "127.0.0.1:7654")
//	-udp string        UDP datagram port (default ":7655")
//	-debug             Enable debug logging
//	-help              Show help message
//
// See SAMv3.md for the complete SAM protocol specification.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-i2p/go-sam-bridge/lib/embedding"
	"github.com/go-i2p/go-sam-bridge/lib/handler"
	"github.com/go-i2p/go-sam-bridge/lib/i2cp"
	"github.com/go-i2p/go-sam-bridge/lib/session"
	samstreaming "github.com/go-i2p/go-sam-bridge/lib/streaming"
	"github.com/go-i2p/go-streaming"
	"github.com/sirupsen/logrus"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"

	// Build info
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	cfg := parseFlags()

	// Configure logging
	log := logrus.New()
	log.SetOutput(os.Stdout)
	if cfg.Debug {
		log.SetLevel(logrus.DebugLevel)
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.WithFields(logrus.Fields{
		"version":   Version,
		"buildTime": BuildTime,
		"commit":    GitCommit,
	}).Info("Starting SAM bridge server")

	// Connect to I2P router for I2CP integration
	i2cpClient, err := connectI2CP(cfg, log)
	if err != nil {
		log.WithError(err).Error("Failed to connect to I2P router")
		log.Info("Make sure I2P is running and SAM interface is enabled")
		os.Exit(1)
	}
	defer i2cpClient.Close()

	log.WithField("version", i2cpClient.RouterVersion()).Info("Connected to I2P router")

	// Create I2CP provider adapter
	i2cpProvider := newI2CPProviderAdapter(i2cpClient)

	// Parse datagram port
	datagramPort := parseDatagramPort(cfg.UDPAddr)

	// Create bridge with embedding API
	bridge, err := embedding.New(
		embedding.WithListenAddr(cfg.ListenAddr),
		embedding.WithI2CPAddr(cfg.I2CPAddr),
		embedding.WithDatagramPort(datagramPort),
		embedding.WithI2CPProvider(i2cpProvider),
		embedding.WithLogger(log),
		embedding.WithDebug(cfg.Debug),
		embedding.WithHandlerRegistrar(createHandlerRegistrar(i2cpClient)),
	)
	if err != nil {
		log.WithError(err).Error("Failed to create bridge")
		os.Exit(1)
	}

	// Start bridge
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := bridge.Start(ctx); err != nil {
		log.WithError(err).Error("Failed to start bridge")
		os.Exit(1)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Received shutdown signal")
	bridge.Stop(context.Background())
}

// Config holds command-line configuration.
type Config struct {
	ListenAddr string
	I2CPAddr   string
	UDPAddr    string
	Debug      bool
	Username   string
	Password   string
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ListenAddr, "listen", ":7656", "SAM listen address")
	flag.StringVar(&cfg.I2CPAddr, "i2cp", "127.0.0.1:7654", "I2CP router address")
	flag.StringVar(&cfg.UDPAddr, "udp", ":7655", "UDP datagram port")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")
	flag.StringVar(&cfg.Username, "user", "", "I2CP username (optional)")
	flag.StringVar(&cfg.Password, "pass", "", "I2CP password (optional)")

	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *showVersion {
		fmt.Printf("sam-bridge %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Println("SAM Bridge - SAMv3.3 Protocol Bridge for I2P")
		fmt.Println()
		fmt.Println("Usage: sam-bridge [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Environment variables:")
		fmt.Println("  SAM_LISTEN    SAM listen address (overrides -listen)")
		fmt.Println("  I2CP_ADDR     I2CP router address (overrides -i2cp)")
		fmt.Println("  SAM_DEBUG     Enable debug logging (overrides -debug)")
		os.Exit(0)
	}

	// Override with environment variables
	if env := os.Getenv("SAM_LISTEN"); env != "" {
		cfg.ListenAddr = env
	}
	if env := os.Getenv("I2CP_ADDR"); env != "" {
		cfg.I2CPAddr = env
	}
	if os.Getenv("SAM_DEBUG") != "" {
		cfg.Debug = true
	}

	return cfg
}

func connectI2CP(cfg *Config, log *logrus.Logger) (*i2cp.Client, error) {
	i2cpConfig := &i2cp.ClientConfig{
		RouterAddr: cfg.I2CPAddr,
		Username:   cfg.Username,
		Password:   cfg.Password,
	}

	client := i2cp.NewClient(i2cpConfig)
	ctx := context.Background()

	log.WithField("addr", cfg.I2CPAddr).Info("Connecting to I2P router")
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func parseDatagramPort(addr string) int {
	if addr == "" {
		return embedding.DefaultDatagramPort
	}
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		portStr = addr
	}
	if port, err := strconv.Atoi(portStr); err == nil {
		return port
	}
	return embedding.DefaultDatagramPort
}

// createHandlerRegistrar returns a custom handler registrar with I2CP integration.
// This extends the default registrar with I2CP-specific session callbacks.
func createHandlerRegistrar(i2cpClient *i2cp.Client) embedding.HandlerRegistrarFunc {
	return func(router *handler.Router, deps *embedding.Dependencies) {
		log := deps.Logger

		// Use default handler registrar for base handlers
		embedding.DefaultHandlerRegistrar()(router, deps)

		// Get the SESSION handler to add I2CP callback
		// The default registrar already created it, we need to extend it
		// For now, we re-register with the extended callback
		streamConnector := handler.NewStreamingConnector()
		streamAcceptor := handler.NewStreamingAcceptor()
		streamForwarder := handler.NewStreamingForwarder()

		sessionHandler := handler.NewSessionHandler(deps.DestManager)
		sessionHandler.SetI2CPProvider(deps.I2CPProvider)

		// Set session created callback for StreamManager wiring
		sessionHandler.SetSessionCreatedCallback(func(sess session.Session, i2cpHandle session.I2CPSessionHandle) {
			if sess.Style() != session.StyleStream || i2cpHandle == nil {
				return
			}

			i2cpSess, ok := i2cpHandle.(*i2cp.I2CPSession)
			if !ok {
				log.WithField("sessionID", sess.ID()).Warn("Cannot create StreamManager: invalid I2CP session type")
				return
			}

			underlyingSession := i2cpSess.Session()
			underlyingClient := i2cpClient.I2CPClient()
			if underlyingSession == nil || underlyingClient == nil {
				log.WithField("sessionID", sess.ID()).Warn("Cannot create StreamManager: no underlying I2CP session/client")
				return
			}

			streamManager, err := streaming.NewStreamManagerFromSession(underlyingClient, underlyingSession)
			if err != nil {
				log.WithField("sessionID", sess.ID()).WithError(err).Warn("Failed to create StreamManager from session")
				return
			}

			adapter, err := samstreaming.NewAdapter(streamManager)
			if err != nil {
				log.WithField("sessionID", sess.ID()).WithError(err).Warn("Failed to create StreamManager adapter")
				return
			}

			streamConnector.RegisterManager(sess.ID(), adapter)
			streamAcceptor.RegisterManager(sess.ID(), adapter)
			streamForwarder.RegisterManager(sess.ID(), adapter)

			log.WithField("sessionID", sess.ID()).Debug("Registered StreamManager for session")
		})

		// Re-register SESSION handlers with extended callback
		router.Register("SESSION CREATE", sessionHandler)
		router.Register("SESSION ADD", sessionHandler)
		router.Register("SESSION REMOVE", sessionHandler)

		// Re-register STREAM handlers with new connectors
		streamHandler := handler.NewStreamHandler(streamConnector, streamAcceptor, streamForwarder)
		router.Register("STREAM CONNECT", streamHandler)
		router.Register("STREAM ACCEPT", streamHandler)
		router.Register("STREAM FORWARD", streamHandler)

		// Wire destination resolver for NAMING handler
		destResolver, err := i2cp.NewClientDestinationResolverAdapter(i2cpClient, 30*time.Second)
		if err == nil {
			namingHandler := handler.NewNamingHandler(deps.DestManager)
			namingHandler.SetDestinationResolver(destResolver)
			router.Register("NAMING LOOKUP", namingHandler)
			log.Debug("Wired destination resolver to NAMING handler")
		}

		log.Debug("Extended handlers with I2CP integration")
	}
}

// i2cpProviderAdapter wraps i2cp.Client to implement session.I2CPSessionProvider.
type i2cpProviderAdapter struct {
	client *i2cp.Client
}

func newI2CPProviderAdapter(client *i2cp.Client) *i2cpProviderAdapter {
	return &i2cpProviderAdapter{client: client}
}

func (a *i2cpProviderAdapter) CreateSessionForSAM(ctx context.Context, samSessionID string, config *session.SessionConfig) (session.I2CPSessionHandle, error) {
	i2cpConfig := &i2cp.SessionConfigFromSession{
		SignatureType:          config.SignatureType,
		EncryptionTypes:        config.EncryptionTypes,
		InboundQuantity:        config.InboundQuantity,
		OutboundQuantity:       config.OutboundQuantity,
		InboundLength:          config.InboundLength,
		OutboundLength:         config.OutboundLength,
		InboundBackupQuantity:  config.InboundBackupQuantity,
		OutboundBackupQuantity: config.OutboundBackupQuantity,
		FastReceive:            config.FastReceive,
		ReduceIdleTime:         config.ReduceIdleTime,
		CloseIdleTime:          config.CloseIdleTime,
	}
	return a.client.CreateSessionForSAM(ctx, samSessionID, i2cpConfig)
}

func (a *i2cpProviderAdapter) IsConnected() bool {
	return a.client.IsConnected()
}

var _ session.I2CPSessionProvider = (*i2cpProviderAdapter)(nil)
