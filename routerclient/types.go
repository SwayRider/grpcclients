package routerclient

import (
	"time"
)

type RouteStep interface {
	Instruction() string
	DistanceInMeters() float64
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
	DistanceInMeters() float64
	Duration() time.Duration
	Steps() []RouteStep
	AddStep(RouteStep)
}

type RouteCtor func(
	distanceInMeters float64,
	duration time.Duration,
) Route

type GeoLocation interface {
	Coordinate() Coordinate
	FeatureType() string
	Label() *string
	Name() *string
	Housenumber() *string
	Street() *string
	Locality() *string
	Postcode() *string
	City() *string
	District() *string
	County() *string
	State() *string
	Country() *string
	CountryCode() *string
	OsmKey() string
	OsmValue() string
	OsmType() string
	OsmId() int64
	Extent() *BoundingBox
	Language() string
}

type GeoLocationCtor func(
	Coordinate,
	string, //featureType,
	*string, // label
	*string, // name
	*string, // housenumber
	*string, // street
	*string, // locality
	*string, // postcode
	*string, // city
	*string, // district
	*string, // county
	*string, // state
	*string, // country
	*string, // country_code
	string, // osm_key
	string, // osm_value
	string, // osm_type
	int64, // osm_id
	*BoundingBox,
	string, // language
) GeoLocation

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

type CountryList struct {
	Countries []string
}

type MapLocation struct {
	MapCenter Coordinate
	Zoom      int
}

type RouteQuery struct {
	From                   Coordinate
	To                     Coordinate
	Vehicle                string
	Algorithm              string
	StandardRoutingOptions StandardRoutingOptions
}

type BoundingBox struct {
	BottomLeft Coordinate
	TopRight   Coordinate
}

type GeocodeQuery struct {
	Query       string
	Tags        []string
	Layers      []string
	Languages   []string
	Limit       *int
	CountryList *CountryList
	MapLocation *MapLocation
}

type ReverseGeocodeQuery struct {
	Location Coordinate
	Tag      *string
	Layers   []string
	Language *string
	Limit    *int32
}
