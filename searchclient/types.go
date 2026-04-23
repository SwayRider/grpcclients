package searchclient

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

type BoundingBox struct {
	BottomLeft Coordinate
	TopRight   Coordinate
}

type SearchResult interface {
	Label() string
	Locality() string
	Region() string
	Country() string
	Confidence() float64
	Layer() string
	Lat() float64
	Lon() float64
	Street() string
	HouseNumber() string
	Id() string
	LocalAdmin() string
	CountryCode() string
	Name() string
}

type SearchResultCtor func(
	label, locality, region, country string,
	confidence float64,
	layer string,
	lat, lon float64,
	street, housenumber, id, localadmin, countryCode, name string,
) SearchResult

type SearchQuery struct {
	Text       string
	Viewport   BoundingBox
	FocusPoint *Coordinate
	Size       int32
	Language   string
}

type ReverseGeocodeQuery struct {
	Point    Coordinate
	Size     int32
	Language string
}
