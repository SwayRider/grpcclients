package healthclient

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	"github.com/swayrider/grpcclients/internal/client"
	healthv1 "github.com/swayrider/protos/health/v1"
)

type Client struct {
	*client.Client[healthv1.HealthServiceClient]
	mux            *sync.Mutex
	getHostAndPort grpcclients.GetHostAndPort
}

func New(getHostAndPort grpcclients.GetHostAndPort) (grpcclients.Client, error) {
	c := &Client{
		mux:            &sync.Mutex{},
		getHostAndPort: getHostAndPort,
	}

	err := c.newConnection()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) CheckConnection() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.Client == nil {
		return c.newConnection()
	}

	err := c.Ping()
	if err != nil {
		c.Close()
		return c.newConnection()
	}
	return nil
}

func (c *Client) newConnection() error {
	host, port := c.getHostAndPort()
	clnt, err := client.New(
		host, port,
		func(conn *grpc.ClientConn) healthv1.HealthServiceClient {
			return healthv1.NewHealthServiceClient(conn)
		},
		Deadline(),
	)
	if err != nil {
		return err
	}

	c.Client = clnt
	return nil
}

func (c Client) Ping() error {
	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err := c.Impl().Ping(ctx, &healthv1.PingRequest{})
	return err
}

func (c Client) Check(
	component string,
) (
	status ServiceStatus,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().Check(ctx, &healthv1.HealthRequest{
		Component: component,
	})
	if err != nil {
		return
	}

	status = ServiceStatus(res.Status)
	return
}
