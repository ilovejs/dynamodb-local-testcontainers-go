package dynamodblocal

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents a DynamoDB Local container
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html
type Container struct {
	testcontainers.Container
}

const (
	image         = "amazon/dynamodb-local:2.2.1"
	targetPort    = nat.Port("8000/tcp")
	containerName = "dynamodb_local"
)

// RunContainer creates an instance of the dynamodb container type
func RunContainer(
	ctx context.Context,
	opts ...testcontainers.ContainerCustomizer,
) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{string(targetPort)},
		WaitingFor:   wait.ForListeningPort(targetPort),
	}

	containerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		err := opt.Customize(&containerReq)
		if err != nil {
			panic(err)
		}
	}

	log.Println("CMD:", containerReq.Cmd)
	log.Println("Image:", containerReq.Image)

	container, err := testcontainers.GenericContainer(ctx, containerReq)
	if err != nil {
		return nil, err
	}

	return &Container{Container: container}, nil
}

// ConnectionString returns DynamoDB local endpoint host and port in <host>:<port> format
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, targetPort)
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}

func (c *Container) GetPort(ctx context.Context) string {
	mappedPort, err := c.MappedPort(ctx, targetPort)
	if err != nil {
		return ""
	}
	return mappedPort.Port()
}

func (c *Container) GetDynamoDBClient(
	ctx context.Context,
) (*dynamodb.Client, error) {
	hostAndPort, err := c.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     "DUMMYIDEXAMPLE",
					SecretAccessKey: "DUMMYEXAMPLEKEY",
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}

	// more option
	dynClient := dynamodb.NewFromConfig(
		cfg,
		dynamodb.WithEndpointResolverV2(
			&DynamoDBLocalResolver{
				hostAndPort: hostAndPort},
		),
	)
	return dynClient, nil
}

// WithSharedDB allows container reuse between successive runs. Data will be persisted.
func WithSharedDB() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if len(req.Cmd) > 0 {
			req.Cmd = append(req.Cmd, "-sharedDb")
		} else {
			req.Cmd = append(req.Cmd, "-jar", "DynamoDBLocal.jar", "-sharedDb")
		}
		req.Name = containerName
		req.Reuse = true
		return nil
	}
}

// WithTelemetryDisabled DynamoDB local will not send any telemetry
func WithTelemetryDisabled() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// if other flags (e.g. -sharedDb) exist, append to them
		if len(req.Cmd) > 0 {
			req.Cmd = append(req.Cmd, "-disableTelemetry")
		} else {
			req.Cmd = append(req.Cmd, "-jar", "DynamoDBLocal.jar", "-disableTelemetry")
		}
		return nil
	}
}

func WithImage(img string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Image = img
		return nil
	}
}
