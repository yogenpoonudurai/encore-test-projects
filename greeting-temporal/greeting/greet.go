package greeting

import (
	"context"
	"encore.app/greeting/workflow"
	"encore.dev/rlog"
	"go.temporal.io/sdk/client"
)

type GreetResponse struct {
	Greeting string
}

//encore:api public path=/greet/:name
func (s *Service) Greet(ctx context.Context, name string) (*GreetResponse, error) {
	options := client.StartWorkflowOptions{
		ID:        "greeting-workflow",
		TaskQueue: greetingTaskQueue,
	}
	we, err := s.client.ExecuteWorkflow(ctx, options, workflow.Greeting, name)
	if err != nil {
		return nil, err
	}
	rlog.Info("started workflow", "id", we.GetID(), "run_id", we.GetRunID())

	// Get the results
	var greeting string
	err = we.Get(ctx, &greeting)
	if err != nil {
		return nil, err
	}
	return &GreetResponse{Greeting: greeting}, nil
}
