package routerclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	"github.com/swayrider/grpcclients/internal/client"
	geo "github.com/swayrider/protos/common_types/geo"
	healthv1 "github.com/swayrider/protos/health/v1"
	routerv1 "github.com/swayrider/protos/router/v1"
)

type Client struct {
	*client.Client[routerv1.RouterServiceClient]
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
		func(conn *grpc.ClientConn) routerv1.RouterServiceClient {
			healthClient = healthv1.NewHealthServiceClient(conn)
			return routerv1.NewRouterServiceClient(conn)
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

func (c Client) Route(
	accessToken string,
	query RouteQuery,
	routeCtor RouteCtor,
	stepCtor RouteStepCtor,
) (
	route Route,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	// Map vehicle string to RoutingMode enum
	mode := vehicleToRoutingMode(query.Vehicle)

	// Build route options from StandardRoutingOptions using preference values
	// A value of 0.0 means "avoid", 1.0 means "prefer"
	routeOptions := &routerv1.RouteOptions{}
	if query.StandardRoutingOptions.AvoidTollRoads {
		tollPref := float64(0)
		routeOptions.TollwayPreference = &tollPref
	}
	if query.StandardRoutingOptions.AvoidHighways {
		hwPref := float64(0)
		routeOptions.HighwayPreference = &hwPref
	}
	if query.StandardRoutingOptions.AvoidFerries {
		ferryPref := float64(0)
		routeOptions.FerryPreference = &ferryPref
	}
	if query.StandardRoutingOptions.ScenicPreference > 0 {
		routeOptions.TrailPreference = &query.StandardRoutingOptions.ScenicPreference
	}
	if query.StandardRoutingOptions.HighwayAvoidance > 0 {
		routeOptions.HighwayPreference = func() *float64 {
			v := 1.0 - query.StandardRoutingOptions.HighwayAvoidance
			return &v
		}()
	}
	if query.StandardRoutingOptions.TollAvoidance > 0 {
		routeOptions.TollwayPreference = func() *float64 {
			v := 1.0 - query.StandardRoutingOptions.TollAvoidance
			return &v
		}()
	}

	// Map motorcycle preferences
	var scenicPref *float32
	var highwayAvoid *float32
	var tollAvoid *float32
	var unpavedHandling *routerv1.UnpavedHandling

	if query.StandardRoutingOptions.ScenicPreference > 0 {
		v := float32(query.StandardRoutingOptions.ScenicPreference)
		scenicPref = &v
	}
	if query.StandardRoutingOptions.HighwayAvoidance > 0 {
		v := float32(query.StandardRoutingOptions.HighwayAvoidance)
		highwayAvoid = &v
	}
	if query.StandardRoutingOptions.TollAvoidance > 0 {
		v := float32(query.StandardRoutingOptions.TollAvoidance)
		tollAvoid = &v
	}
	if query.StandardRoutingOptions.UnpavedHandling != "" {
		v := unpdavedHandlingToProto(query.StandardRoutingOptions.UnpavedHandling)
		unpavedHandling = &v
	}

	res, err := c.Impl().Route(ctx, &routerv1.RouteRequest{
		Mode: mode,
		Locations: []*routerv1.RouteLocation{
			{
				Location: &geo.Coordinate{
					Lat: query.From.Latitude,
					Lon: query.From.Longitude,
				},
				Type: routerv1.LocationType_L_BREAK,
			},
			{
				Location: &geo.Coordinate{
					Lat: query.To.Latitude,
					Lon: query.To.Longitude,
				},
				Type: routerv1.LocationType_L_BREAK,
			},
		},
		RouteOptions:     routeOptions,
		ScenicPreference: scenicPref,
		HighwayAvoidance: highwayAvoid,
		TollAvoidance:    tollAvoid,
		UnpavedHandling:  unpavedHandling,
	})
	if err != nil {
		return
	}

	if res.Trip == nil || res.Trip.Summary == nil {
		err = errors.New("invalid route response: missing trip or summary")
		return
	}

	// Extract distance (in meters) and duration (in seconds) from trip summary
	route = routeCtor(
		res.Trip.Summary.Length,
		time.Duration(res.Trip.Summary.Time)*time.Second,
	)

	// Extract steps from maneuvers in each leg
	for _, leg := range res.Trip.Legs {
		for _, maneuver := range leg.Maneuvers {
			route.AddStep(stepCtor(
				maneuver.Instruction,
				maneuver.Length,
				time.Duration(maneuver.Time)*time.Second,
				Coordinate{
					Latitude:  0, // Maneuvers don't have direct coordinates
					Longitude: 0,
				},
			))
		}
	}
	return
}

// vehicleToRoutingMode maps vehicle string to RoutingMode enum
func vehicleToRoutingMode(vehicle string) routerv1.RoutingMode {
	switch vehicle {
	case "motorscooter":
		return routerv1.RoutingMode_RM_MOTORSCOOTER
	case "motorcycle":
		return routerv1.RoutingMode_RM_MOTORCYCLE
	default:
		return routerv1.RoutingMode_RM_CAR
	}
}

// Note: GeoCode functionality has been removed as the Geocode RPC
// is no longer part of the RouterService proto definition.

// unpdavedHandlingToProto maps string to UnpavedHandling enum
func unpdavedHandlingToProto(h string) routerv1.UnpavedHandling {
	switch h {
	case "prefer":
		return routerv1.UnpavedHandling_UH_PREFER
	case "avoid":
		return routerv1.UnpavedHandling_UH_AVOID
	default:
		return routerv1.UnpavedHandling_UH_NEUTRAL
	}
}
