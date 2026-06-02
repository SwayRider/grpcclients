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
}
