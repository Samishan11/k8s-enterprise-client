package k8sclient

import "time"

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
