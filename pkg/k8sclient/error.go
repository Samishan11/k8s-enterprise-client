package k8sclient

import "github.com/pkg/errors"

var (
	ErrLeaderElectionNotConfigured = errors.New("leader election not configured")
	ErrNotLeader                   = errors.New("current instance is not the leader")
	ErrCacheMiss                   = errors.New("cache miss")
	ErrInvalidConfiguration        = errors.New("invalid configuration")
)

type LeaderElectionError struct {
	Err error
}

func (e *LeaderElectionError) Error() string {
	return "leader election error: " + e.Err.Error()
}
