package k8sclient

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
