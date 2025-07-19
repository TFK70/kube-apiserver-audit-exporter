package server

import (
	"fmt"

	"github.com/TFK70/kube-apiserver-audit-exporter/internal/logging"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Server struct {
	BindPort int
	Handlers map[string]echo.HandlerFunc

	logger *logrus.Entry
}

type ServerOptions struct {
	BindPort int
	Handlers map[string]echo.HandlerFunc
}

type Option func(*ServerOptions)

func WithBindPort(port int) Option {
	return func(o *ServerOptions) {
		o.BindPort = port
	}
}

func WithHander(path string, handler echo.HandlerFunc) Option {
	return func(o *ServerOptions) {
		o.Handlers[path] = handler
	}
}

func NewServer(opts ...Option) (*Server, error) {
	options := &ServerOptions{
		Handlers: make(map[string]echo.HandlerFunc),
	}

	for _, opt := range opts {
		opt(options)
	}

	logger, err := logging.GetNamedLogger("server.go")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}
	logger = logger.WithFields(logrus.Fields{
		"bindPort": options.BindPort,
	})

	logger.Infof("Initialized server")

	return &Server{
		BindPort: options.BindPort,
		Handlers: options.Handlers,
		logger:   logger,
	}, nil
}

func (s *Server) Start() error {
	e := echo.New()

	for path, handler := range s.Handlers {
		e.Any(path, handler)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", s.BindPort)))

	return nil
}
