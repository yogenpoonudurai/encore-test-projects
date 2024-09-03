package greeting

import (
	"context"
	"encore.app/greeting/workflow"
	"fmt"

	"encore.dev"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Use an environment-specific task queue so we can use the same
// Temporal Cluster for all cloud environments.
var (
	envName           = encore.Meta().Environment.Name
	greetingTaskQueue = envName + "-greeting"
)

//encore:service
type Service struct {
	client client.Client
	worker worker.Worker
}

func initService() (*Service, error) {
	c, err := client.Dial(client.Options{})

	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}

	w := worker.New(c, greetingTaskQueue, worker.Options{})

	w.RegisterWorkflow(workflow.Greeting)
	w.RegisterActivity(workflow.ComposeGreeting)

	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %v", err)
	}
	return &Service{client: c, worker: w}, nil
}

func (s *Service) Shutdown(force context.Context) {
	s.client.Close()
	s.worker.Stop()
}
