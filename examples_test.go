package dynamodblocal

import (
	"context"
	"fmt"
	"log"
)

func ExampleRunContainer() {
	ctx := context.Background()

	dynamodbContainer, err := RunContainer(
		ctx,
		WithTelemetryDisabled(),
		WithSharedDB(),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := dynamodbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	state, err := dynamodbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
