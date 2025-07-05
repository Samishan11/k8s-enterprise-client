package main

import (
	"context"
	"log"

	"k8s-enterprise-client/pkg/k8sclient"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/labels"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	opts := k8sclient.DefaultOptions()
	opts.KubeconfigPath = ""
	opts.EnableLeaderElection = true
	opts.LeaderElectionID = "k8s-client-leader"

	client, err := k8sclient.NewClient(context.Background(), opts, logger)
	if err != nil {
		logger.Fatal("Failed to create client", zap.Error(err))
	}

	pods, err := client.ListPodsEnhanced(context.Background(), "default", labels.Everything(), true)
	if err != nil {
		logger.Error("Failed to list pods", zap.Error(err))
	}

	logger.Info("Found pods", zap.Int("count", len(pods)))
}
