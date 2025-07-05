package k8sclient

import "time"

func DefaultOptions() ClientOptions {
	return ClientOptions{
		QPS:                  50,
		Burst:                100,
		Timeout:              30 * time.Second,
		CacheTTL:             5 * time.Minute,
		CacheCleanupInterval: 10 * time.Minute,
		LeaderElectionNS:     "default",
	}
}
