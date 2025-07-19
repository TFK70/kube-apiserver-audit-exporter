package collector

import (
	"fmt"

	vm "github.com/VictoriaMetrics/metrics"
	"github.com/sirupsen/logrus"

	"github.com/TFK70/kube-apiserver-audit-exporter/internal/audit"
	"github.com/TFK70/kube-apiserver-audit-exporter/internal/logging"
)

type APIServerRequestsCollector struct {
	AuditLogPath string

	reader *audit.AuditReader
	logger *logrus.Entry
}

type APIServerRequestsCollectorOptions struct {
	AuditLogPath string
}

type Option func(*APIServerRequestsCollectorOptions)

func WithAuditLogPath(path string) Option {
	return func(o *APIServerRequestsCollectorOptions) {
		o.AuditLogPath = path
	}
}

func NewAPIServerRequestsCollector(opts ...Option) (*APIServerRequestsCollector, error) {
	options := &APIServerRequestsCollectorOptions{}

	for _, opt := range opts {
		opt(options)
	}

	logger, err := logging.GetNamedLogger("collector.go")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}
	logger = logger.WithFields(logrus.Fields{
		"auditLogPath": options.AuditLogPath,
	})

	logger.Infof("Initialized collector")

	return &APIServerRequestsCollector{
		AuditLogPath: options.AuditLogPath,

		logger: logger,
	}, nil
}

func (c *APIServerRequestsCollector) Collect(event *audit.Event) {
	counter := vm.GetOrCreateCounter(
		fmt.Sprintf(
			`apiserver_requests{resource="%s",resourceName="%s",resourceNamespace="%s",resourceApiGroup="%s",resourceApiVersion="%s",verb="%s",username="%s",userAgent="%s",responseCode="%d"}`,
			event.ObjectRef.Resource,
			event.ObjectRef.Name,
			event.ObjectRef.Namespace,
			event.ObjectRef.APIGroup,
			event.ObjectRef.APIVersion,
			event.Verb,
			event.User.Username,
			event.UserAgent,
			event.ResponseStatus.Code,
		),
	)
	counter.Inc()
}

func (c *APIServerRequestsCollector) Start() error {
	var err error
	c.reader, err = audit.NewReader(
		audit.WithPath(c.AuditLogPath),
	)
	if err != nil {
		return fmt.Errorf("failed to create reader: %v", err)
	}

	err = c.reader.Start()
	if err != nil {
		return fmt.Errorf("failed to start reader: %v", err)
	}

	go func() {
		for event := range c.reader.Events {
			c.Collect(&event)
		}
	}()

	return nil
}

func (c *APIServerRequestsCollector) Stop() error {
	err := c.reader.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop reader: %v", err)
	}

	vm.UnregisterAllMetrics()

	return nil
}
