package searchclient

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	"github.com/swayrider/grpcclients/internal/client"
	geo "github.com/swayrider/protos/common_types/geo"
	healthv1 "github.com/swayrider/protos/health/v1"
	searchv1 "github.com/swayrider/protos/search/v1"
)

type Client struct {
	*client.Client[searchv1.SearchServiceClient]
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
		func(conn *grpc.ClientConn) searchv1.SearchServiceClient {
			healthClient = healthv1.NewHealthServiceClient(conn)
			return searchv1.NewSearchServiceClient(conn)
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

func (c *Client) Ping() error {
	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err := c.healthClient.Ping(ctx, &healthv1.PingRequest{})
	return err
}

func (c *Client) Search(
	accessToken string,
	query SearchQuery,
	ctor SearchResultCtor,
) (
	results []SearchResult,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	req := &searchv1.SearchRequest{
		Text: query.Text,
		Viewport: &geo.BoundingBox{
			BottomLeft: &geo.Coordinate{
				Lat: query.Viewport.BottomLeft.Latitude,
				Lon: query.Viewport.BottomLeft.Longitude,
			},
			TopRight: &geo.Coordinate{
				Lat: query.Viewport.TopRight.Latitude,
				Lon: query.Viewport.TopRight.Longitude,
			},
		},
	}
	if query.FocusPoint != nil {
		req.FocusPoint = &geo.Coordinate{
			Lat: query.FocusPoint.Latitude,
			Lon: query.FocusPoint.Longitude,
		}
	}
	if query.Size > 0 {
		req.Size = &query.Size
	}
	if query.Language != "" {
		req.Language = &query.Language
	}

	res, err := c.Impl().Search(ctx, req)
	if err != nil {
		return
	}

	results = make([]SearchResult, 0, len(res.Results))
	for _, r := range res.Results {
		results = append(results, ctor(
			r.Label, r.Locality, r.Region, r.Country,
			r.Confidence,
			r.Layer,
			r.Lat, r.Lon,
			r.Street, r.Housenumber, r.Id, r.Localadmin, r.CountryCode, r.Name,
		))
	}
	return
}

func (c *Client) ReverseGeocode(
	accessToken string,
	query ReverseGeocodeQuery,
	ctor SearchResultCtor,
) (
	results []SearchResult,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	req := &searchv1.ReverseGeocodeRequest{
		Point: &geo.Coordinate{
			Lat: query.Point.Latitude,
			Lon: query.Point.Longitude,
		},
	}
	if query.Size > 0 {
		req.Size = &query.Size
	}
	if query.Language != "" {
		req.Language = &query.Language
	}

	res, err := c.Impl().ReverseGeocode(ctx, req)
	if err != nil {
		return
	}

	results = make([]SearchResult, 0, len(res.Results))
	for _, r := range res.Results {
		results = append(results, ctor(
			r.Label, r.Locality, r.Region, r.Country,
			r.Confidence,
			r.Layer,
			r.Lat, r.Lon,
			r.Street, r.Housenumber, r.Id, r.Localadmin, r.CountryCode, r.Name,
		))
	}
	return
}
