package k8sclient

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	GetConfig() *rest.Config
	GetClientset() kubernetes.Interface
	ListPodsEnhanced(ctx context.Context, namespace string, labelSelector labels.Selector, includeMetrics bool) ([]corev1.Pod, error)
	WatchNamespaces(ctx context.Context, eventHandler func(eventType string, ns *corev1.Namespace)) error
	IsLeader() bool
	RunLeaderElection(ctx context.Context) error
}
