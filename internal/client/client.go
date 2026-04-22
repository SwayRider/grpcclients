package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client[T any] struct {
	conn     *grpc.ClientConn
	client   T
	deadline time.Duration
}

func New[T any](
	host string,
	port int,
	ctor func(*grpc.ClientConn) T,
	deadline time.Duration,
) (*Client[T], error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client[T]{
		conn:     conn,
		client:   ctor(conn),
		deadline: deadline,
	}, nil
}

func (c *Client[T]) Close() error {
	return c.conn.Close()
}

func (c Client[T]) Impl() T {
	return c.client
}

func (c Client[T]) Deadline() time.Duration {
	return c.deadline
}

func (c *Client[T]) SetDeadline(deadline time.Duration) {
	c.deadline = deadline
}

func (c Client[T]) Context(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.deadline)
}

func (c Client[T]) AuthorizedContext(
	ctx context.Context,
	jwt string,
) (context.Context, context.CancelFunc) {
	md := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", jwt))
	return c.Context(metadata.NewOutgoingContext(ctx, md))
}
