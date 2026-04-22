package mailclient

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	"github.com/swayrider/grpcclients/internal/client"
	healthv1 "github.com/swayrider/protos/health/v1"
	mailv1 "github.com/swayrider/protos/mail/v1"
)

type Client struct {
	*client.Client[mailv1.MailServiceClient]
	healthClient   healthv1.HealthServiceClient
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
	var healthClient healthv1.HealthServiceClient
	clnt, err := client.New(
		host, port,
		func(conn *grpc.ClientConn) mailv1.MailServiceClient {
			healthClient = healthv1.NewHealthServiceClient(conn)
			return mailv1.NewMailServiceClient(conn)
		},
		Deadline(),
	)
	if err != nil {
		return err
	}

	c.Client = clnt
	c.healthClient = healthClient
	return nil
}

func (c Client) Ping() error {
	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err := c.healthClient.Ping(ctx, &healthv1.PingRequest{})
	return err
}

func (c Client) SendTemplate(
	accessToken string,
	mail *TemplateMail,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().SendTemplate(ctx, mail.Request())
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) SendTemplateInternal(
	mail *TemplateMail,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().SendTemplateInternal(ctx, mail.Request())
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) Send(
	accessToken string,
	mail *Mail,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().Send(ctx, mail.Request())
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) SendInternal(
	mail *Mail,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().SendInternal(ctx, mail.Request())
	if err != nil {
		return
	}

	message = res.Message
	return
}
