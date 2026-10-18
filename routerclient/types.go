package routerclient

import (
	"time"
)

type RouteStep interface {
	Instruction() string
	DistanceKm() float64
	Duration() time.Duration
	Coordinate() Coordinate
}

type RouteStepCtor func(
	instruction string,
	distanceInMeters float64,
	duration time.Duration,
	coordinate Coordinate,
) RouteStep

type Route interface {
	DistanceKm() float64
	Duration() time.Duration
	Steps() []RouteStep
	AddStep(RouteStep)
}

type RouteCtor func(
	distanceInMeters float64,
	duration time.Duration,
) Route

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

type StandardRoutingOptions struct {
	AvoidTollRoads   bool
	AvoidHighways    bool
	AvoidFerries     bool
	FuelEfficient    bool
	ScenicPreference float64
	HighwayAvoidance float64
	TollAvoidance    float64
	UnpavedHandling  string
}

type Waypoint struct {
	Coordinate Coordinate
	Type       string // "break" or "through"
}

type RouteQuery struct {
	From                   Coordinate
	To                     Coordinate
	Waypoints              []Waypoint
	Vehicle                string
	Algorithm              string
	StandardRoutingOptions StandardRoutingOptions
	Unit                   string // "km" or "mi"
	Language               string

	// FromHeading is the direction of travel at From in degrees clockwise
	// from north; the router prefers roads that go that way. Used when a
	// rider leaves a route, so the new route does not start with a U-turn.
	FromHeading *int
	// FromHeadingTolerance is how far a road may deviate from FromHeading,
	// in degrees. Router default when nil.
	FromHeadingTolerance *int
	// ExcludePolygons are areas the route must not pass through.
	ExcludePolygons [][]Coordinate
}

// RejoinPoint is a point of the original route a rider can rejoin at.
type RejoinPoint struct {
	Coordinate Coordinate
	// Heading is the direction of the original route at the point in
	// degrees clockwise from north.
	Heading int
	// RemainingSeconds is the travel time left on the original route after
	// the point.
	RemainingSeconds float64
}

// RejoinQuery asks for the way back onto a route: Fresh describes a fresh
// route from the rider's position through the remaining stops (with its
// heading, options and exclusions), Points are the candidate rejoin points.
type RejoinQuery struct {
	Fresh  RouteQuery
	Points []RejoinPoint
}
