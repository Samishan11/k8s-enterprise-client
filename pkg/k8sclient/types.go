package k8sclient

import (
	"sync"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/Samishan11/k8s-enterprise-client/internal/cache"
)

type ResourceMetrics struct {
	CPU    string
	Memory string
}

type EnhancedPod struct {
	corev1.Pod
	Metrics *ResourceMetrics
}

type WatchEvent struct {
	Type   string
	Object runtime.Object
}

type ClientOptions struct {
	KubeconfigPath       string
	QPS                  float32
	Burst                int
	Timeout              time.Duration
	UserAgent            string
	EnableLeaderElection bool
	LeaderElectionID     string
	LeaderElectionNS     string
	CacheTTL             time.Duration
	CacheCleanupInterval time.Duration
}

type Client struct {
	Clientset     kubernetes.Interface
	DynamicClient dynamic.Interface
	MetricsClient *versioned.Clientset
	config        *rest.Config
	logger        *zap.Logger
	leaderElector *leaderelection.LeaderElector
	isLeader      bool
	leaderMutex   sync.RWMutex
	cache         *cache.ResourceCache
}
