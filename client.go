package grpcclients

type GetHostAndPort func() (string, int)

type Client interface {
	CheckConnection() error
	Close() error
}
