# Kubernetes Enterprise Client

![Go Version](https://img.shields.io/github/go-mod/go-version/Samishan11/k8s-enterprise-client)
![Release](https://img.shields.io/github/v/release/Samishan11/k8s-enterprise-client)
![License](https://img.shields.io/badge/license-Apache--2.0-blue)

A production-grade Go client for Kubernetes with enterprise features including intelligent caching, leader election, and real-time resource monitoring.

## Installation

- Go 1.21+
- Kubernetes cluster (or kubectl configured)
- kubectl access (for some examples)

### Install the Client Library

````bash
# Install the latest stable version
go get github.com/Samishan11/k8s-enterprise-client@latest

# Or pin to specific version
go get github.com/Samishan11/k8s-enterprise-client@v1.24.2.


### Verify Installation
Create test file `main.go`:
```go
package main

import (
	"fmt"
	k8s "github.com/Samishan11/k8s-enterprise-client/pkg/k8sclient"
)

func main() {
	fmt.Println("Client version:", k8s.Version)
}
````

Then run:

```bash
go run main.go
```

## Basic Usage Example

### Initialize Client

```go
package main

import (
	"context"
	"log"

	"go.uber.org/zap"
	k8s "github.com/Samishan11/k8s-enterprise-client/pkg/k8sclient"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	client, err := k8s.NewClient(
		context.Background(),
		k8s.DefaultOptions(),
		logger,
	)
	if err != nil {
		logger.Fatal("Client init failed", zap.Error(err))
	}

	logger.Info("Client initialized successfully")
}
```

## Advanced Examples

### Example 1: List Pods with Metrics

```go
func listPodsWithMetrics(client *k8s.Client, namespace string) {
	pods, err := client.ListPodsEnhanced(
		context.Background(),
		namespace,
		nil,
		true,
	)
	if err != nil {
		client.Logger.Error("Failed to list pods", zap.Error(err))
		return
	}

	for _, pod := range pods {
		client.Logger.Info("Pod",
			zap.String("name", pod.Name),
			zap.String("status", string(pod.Status.Phase)),
			zap.String("cpu", pod.Annotations["metrics/cpu"]),
			zap.String("memory", pod.Annotations["metrics/memory"]),
		)
	}
}
```

### Example 2: Cached Resource Watch

```go
func watchWithCache(client *k8s.Client, resourceType, namespace string) {
	cacheKey := fmt.Sprintf("%s-%s", resourceType, namespace)

	// Check cache first
	if cached, exists := client.Cache.Get(cacheKey); exists {
		client.Logger.Info("Using cached data",
			zap.String("key", cacheKey))
		return
	}

	// Set up watch
	err := client.WatchResources(
		context.Background(),
		resourceType,
		namespace,
		func(eventType string, obj interface{}) {
			client.Logger.Info("Resource event",
				zap.String("type", eventType),
				zap.String("resource", fmt.Sprintf("%T", obj)))

			// Update cache on modifications
			if eventType == "MODIFIED" {
				client.Cache.Set(cacheKey, obj, 10*time.Minute)
			}
		})
	if err != nil {
		client.Logger.Error("Watch failed",
			zap.String("resource", resourceType),
			zap.Error(err))
	}
}
```

## Configuration Options

### Available Options

```go
type ClientOptions struct {
	KubeconfigPath       string
	QPS                  float32
	Burst                int
	Timeout              time.Duration
	EnableLeaderElection bool
	LeaderElectionID     string
	CacheTTL             time.Duration
}
```

### Recommended Production Settings

```go
opts := k8s.DefaultOptions()
opts.QPS = 100
opts.Burst = 200
opts.EnableLeaderElection = true
opts.LeaderElectionID = "my-app-leader"
```

## Code Formatting Tips

1. Use triple backticks (```) to start/end code blocks
2. Specify language after opening backticks for syntax highlighting:
   ```go
   // Go code
   ```
   ```bash
   # Shell commands
   ```
3. Keep examples concise but functional
4. Include relevant imports in each example
5. Show error handling patterns

## Best Practices

1. **For long examples** (>20 lines), consider linking to example files:
   ```markdown
   See [examples/pod_monitor.go](./examples/pod_monitor.go) for complete implementation.
   ```
2. **For complex workflows** use numbered steps with embedded code:

   ````markdown
   1. First initialize the client:
      ```go
      client, err := k8s.NewClient(...)
      ```
   ````

   2. Then set up watches:
      ```go
      err = client.WatchResources(...)
      ```

   ```

   ```

3. **For CLI examples** show both command and expected output:
   ```bash
   $ go run main.go
   ```
