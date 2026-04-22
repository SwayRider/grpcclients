package regionclient

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	"github.com/swayrider/grpcclients/internal/client"
	"github.com/swayrider/protos/common_types/geo"
	healthv1 "github.com/swayrider/protos/health/v1"
	regionv1 "github.com/swayrider/protos/region/v1"
)

type Client struct {
	*client.Client[regionv1.RegionServiceClient]
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
		func(conn *grpc.ClientConn) regionv1.RegionServiceClient {
			healthClient = healthv1.NewHealthServiceClient(conn)
			return regionv1.NewRegionServiceClient(conn)
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

func (c *Client) SearchPoint(
	location Coordinate,
	includeExtended bool,
) (
	regions RegionList,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().SearchPoint(ctx, &regionv1.SearchPointRequest{
		Location: &geo.Coordinate{
			Lat: location.Latitude,
			Lon: location.Longitude,
		},
		IncludeExtended: includeExtended,
	})
	if err != nil {
		return
	}

	regions.CoreRegions = res.CoreRegions
	regions.ExtendedRegions = res.ExtendedRegions
	return
}

func (c *Client) SearchBox(
	boundingBox BoundingBox,
	includeExtended bool,
) (
	regions RegionList,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().SearchBox(ctx, &regionv1.SearchBoxRequest{
		Box: &geo.BoundingBox{
			BottomLeft: &geo.Coordinate{
				Lat: boundingBox.BottomLeft.Latitude,
				Lon: boundingBox.BottomLeft.Longitude,
			},
			TopRight: &geo.Coordinate{
				Lat: boundingBox.TopRight.Latitude,
				Lon: boundingBox.TopRight.Longitude,
			},
		},
		IncludeExtended: includeExtended,
	})
	if err != nil {
		return
	}

	regions.CoreRegions = res.CoreRegions
	regions.ExtendedRegions = res.ExtendedRegions
	return
}

func (c *Client) SearchRadius(
	location Coordinate,
	radiusKm float64,
	includeExtended bool,
) (
	regions RegionList,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().SearchRadius(ctx, &regionv1.SearchRadiusRequest{
		Location: &geo.Coordinate{
			Lat: location.Latitude,
			Lon: location.Longitude,
		},
		RadiusKm:        radiusKm,
		IncludeExtended: includeExtended,
	})
	if err != nil {
		return
	}

	regions.CoreRegions = res.CoreRegions
	regions.ExtendedRegions = res.ExtendedRegions
	return
}

func (c *Client) FindCrossingLocations(
	fromRegion, toRegion string,
	flomLocation, toLocation Coordinate,
	config BorderCrossingConfig,
	limit int,
) (
	crossings []BorderCrossing,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	var req *regionv1.FindCrossingLocationsRequest

	if config.Type() == "simple" {
		cfg := config.(BorderCrossingSimpleConfig)
		rtOrder := make([]regionv1.RoadType, 0, len(cfg.RoadTypeOrder))
		for _, rt := range cfg.RoadTypeOrder {
			rtOrder = append(rtOrder, regionv1.RoadType(rt.ToCode()))
		}
		req = &regionv1.FindCrossingLocationsRequest{
			FromRegion:       fromRegion,
			ToRegion:         toRegion,
			FromLocation:	  &geo.Coordinate{Lat: flomLocation.Latitude, Lon: flomLocation.Longitude},
			ToLocation:       &geo.Coordinate{Lat: toLocation.Latitude, Lon: toLocation.Longitude},
			ConfigOneof:      &regionv1.FindCrossingLocationsRequest_SimpleConfig{
				SimpleConfig: &regionv1.BorderCrossingSimpleConfig{
					RoadTypeOrder: rtOrder, 
					RoadTypeDelta: cfg.RoadTypeDelta,
				
			}},
			Limit:            int32(limit),
		}
	} else if config.Type() == "advanced" {
		cfg := config.(BorderCrossingAdvancedConfig)

		defLst := make([]*regionv1.BorderCrossingDefinition, 0, len(cfg.Definitions))
		for _, def := range cfg.Definitions {
			rtOrder := make([]regionv1.RoadType, 0, len(def.RoadTypeOrder))
			for _, rt := range def.RoadTypeOrder {
				rtOrder = append(rtOrder, regionv1.RoadType(rt.ToCode()))
			}
			defLst = append(defLst, &regionv1.BorderCrossingDefinition{
				MaxBorderDistance: def.MaxBorderDistance,
				RoadTypeOrder: rtOrder,
				RoadTypeDelta: def.RoadTypeDelta,
				DropDistance: def.DropDistance,
			})
		}
		req = &regionv1.FindCrossingLocationsRequest{
			FromRegion:       fromRegion,
			ToRegion:         toRegion,
			FromLocation:	  &geo.Coordinate{Lat: flomLocation.Latitude, Lon: flomLocation.Longitude},
			ToLocation:       &geo.Coordinate{Lat: toLocation.Latitude, Lon: toLocation.Longitude},
			ConfigOneof:      &regionv1.FindCrossingLocationsRequest_AdvancedConfig{
				AdvancedConfig: &regionv1.BorderCrossingAdvancedConfig{
					Definitions: defLst,
				},
			},
			Limit:            int32(limit),
		}
		
	} else {
		err = fmt.Errorf("unknown config type: %s", config.Type())
		return
	}

	res, err := c.Impl().FindCrossingLocations(ctx, req)
	if err != nil {
		return
	}

	crossing := make([]BorderCrossing, 0, len(res.Crossings))
	for _, c := range res.Crossings {
		crossing = append(crossing, BorderCrossing{
			FromRegion: c.FromRegion,
			ToRegion:   c.ToRegion,
			RoadType:   ToRoadType(int32(c.RoadType)),
			Location: Coordinate{
				Latitude:  c.Location.Lat,
				Longitude: c.Location.Lon,
			},
			OsmID: int(c.OsmId),
		})
	}
	crossings = crossing
	return
}

func (c *Client) FindRegionPath(
	fromRegion, toRegion string,
) (
	path []string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().FindRegionPath(ctx, &regionv1.FindRegionPathRequest{
		FromRegion: fromRegion,
		ToRegion:   toRegion,
	})
	if err != nil {
		return
	}

	path = res.Path
	return
}

