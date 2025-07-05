package leader

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type Elector struct {
	isLeader bool
	mu       sync.RWMutex
	logger   *zap.Logger
}

func NewElector(clientset kubernetes.Interface, lockName, namespace, identity string, logger *zap.Logger) (*Elector, error) {
	return &Elector{
		logger: logger,
	}, nil
}

func (e *Elector) Run(ctx context.Context) error {
	return nil
}

func (e *Elector) IsLeader() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isLeader
}
