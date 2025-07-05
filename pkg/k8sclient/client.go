package k8sclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// NewClient creates a new enterprise-grade Kubernetes client
func NewClient(ctx context.Context, opts ClientOptions, logger *zap.Logger) (*Client, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (for running inside Kubernetes)
	if config, err = rest.InClusterConfig(); err != nil {
		// Fall back to kubeconfig
		kubeconfigPath := opts.KubeconfigPath
		if kubeconfigPath == "" {
			kubeconfigPath = filepath.Join(homeDir(), ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kubeconfig")
		}
	}

	// Configure client settings
	if opts.QPS > 0 {
		config.QPS = opts.QPS
	}
	if opts.Burst > 0 {
		config.Burst = opts.Burst
	}
	if opts.Timeout > 0 {
		config.Timeout = opts.Timeout
	}
	if opts.UserAgent != "" {
		config.UserAgent = opts.UserAgent
	}

	// Create clients
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset")
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dynamic client")
	}

	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create metrics client")
	}

	client := &Client{
		Clientset:     clientset,
		DynamicClient: dynamicClient,
		MetricsClient: metricsClient,
		config:        config,
		logger:        logger,
		cache:         NewResourceCache(5*time.Minute, 10*time.Minute),
	}

	// Set up leader election if enabled
	if opts.EnableLeaderElection {
		if err := client.setupLeaderElection(ctx, opts.LeaderElectionID); err != nil {
			return nil, errors.Wrap(err, "failed to set up leader election")
		}
	}

	return client, nil
}

// setupLeaderElection configures leader election for HA deployments
func (c *Client) setupLeaderElection(ctx context.Context, lockName string) error {
	if lockName == "" {
		return errors.New("leader election ID cannot be empty")
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockName,
			Namespace: "default",
		},
		Client: c.Clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: os.Getenv("POD_NAME"), // Typically set via downward API
		},
	}

	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				c.leaderMutex.Lock()
				defer c.leaderMutex.Unlock()
				c.isLeader = true
				c.logger.Info("Now acting as leader")
			},
			OnStoppedLeading: func() {
				c.leaderMutex.Lock()
				defer c.leaderMutex.Unlock()
				c.isLeader = false
				c.logger.Info("No longer leader")
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create leader elector")
	}

	c.leaderElector = elector
	go elector.Run(ctx)
	return nil
}

// IsLeader returns true if this instance is the current leader
func (c *Client) IsLeader() bool {
	c.leaderMutex.RLock()
	defer c.leaderMutex.RUnlock()
	return c.isLeader
}

// ListPodsEnhanced lists pods with advanced filtering and metrics
func (c *Client) ListPodsEnhanced(ctx context.Context, namespace string, labelSelector labels.Selector, includeMetrics bool) ([]corev1.Pod, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("pods-%s-%s", namespace, labelSelector.String())
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.([]corev1.Pod), nil
	}

	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	// Get metrics if requested
	if includeMetrics {
		podMetrics, err := c.MetricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
		if err != nil {
			c.logger.Warn("Failed to get pod metrics", zap.Error(err))
		} else {
			// Enrich pod data with metrics
			for i, pod := range pods.Items {
				for _, metric := range podMetrics.Items {
					if pod.Name == metric.Name && pod.Namespace == metric.Namespace {
						pods.Items[i].Annotations["metrics/cpu"] = metric.Containers[0].Usage.Cpu().String()
						pods.Items[i].Annotations["metrics/memory"] = metric.Containers[0].Usage.Memory().String()
					}
				}
			}
		}
	}

	// Update cache
	c.cache.Set(cacheKey, pods.Items, 0) // Use default TTL

	return pods.Items, nil
}

// WatchNamespaces sets up a watch on namespaces with event handling
func (c *Client) WatchNamespaces(ctx context.Context, eventHandler func(eventType string, ns *corev1.Namespace)) error {
	watcher, err := c.Clientset.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to watch namespaces")
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.ResultChan():
				if !ok {
					c.logger.Warn("Namespace watch channel closed")
					return
				}
				if ns, ok := event.Object.(*corev1.Namespace); ok {
					eventHandler(string(event.Type), ns)
				}
			case <-ctx.Done():
				watcher.Stop()
				return
			}
		}
	}()

	return nil
}

// GetResourceUsage returns resource usage metrics for the cluster
func (c *Client) GetResourceUsage(ctx context.Context) (map[string]interface{}, error) {
	// Implementation would collect node metrics, pod metrics, etc.
	// Return aggregated view of cluster resource usage
	return nil, nil
}

// Helper to get home directory for default kubeconfig path
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows fallback
}
