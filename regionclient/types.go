package regionclient

type Coordinate struct {
	Latitude float64
	Longitude float64
}

type BoundingBox struct {
	BottomLeft Coordinate
	TopRight Coordinate
}

type RegionList struct {
	CoreRegions []string
	ExtendedRegions []string
}

type RoadType string

const (
	RT_MOTORWAY RoadType = "motorway"
	RT_TRUNK RoadType = "trunk"
	RT_PRIMARY RoadType = "primary"
	RT_SECONDARY RoadType = "secondary"
)

func ToRoadType(code int32) RoadType {
	switch code {
		case 0:
			return RT_MOTORWAY
		case 1:
			return RT_TRUNK
		case 2:
			return RT_PRIMARY
		case 3:
			return RT_SECONDARY
		default:
			return RT_MOTORWAY
	}
}

func (rt RoadType) ToCode() int32 {
	switch rt {
		case RT_MOTORWAY:
			return 0
		case RT_TRUNK:
			return 1
		case RT_PRIMARY:
			return 2
		case RT_SECONDARY:
			return 3
		default:
			return 0
	}
}

type BorderCrossing struct {
	FromRegion string
	ToRegion string
	RoadType RoadType
	Location Coordinate
	OsmID int
}

type BorderCrossingConfig interface {
	Type() string
}

type BorderCrossingSimpleConfig struct {
	RoadTypeOrder []RoadType
	RoadTypeDelta float64
	DropDistance  float64
}

func (c BorderCrossingSimpleConfig) Type() string {
	return "simple"
}

type BorderCrossingDefinition struct {
	MaxBorderDistance float64
	RoadTypeOrder []RoadType
	RoadTypeDelta float64
	DropDistance float64
}

type BorderCrossingAdvancedConfig struct {
	Definitions []BorderCrossingDefinition
}

func (c BorderCrossingAdvancedConfig) Type() string {
	return "advanced"
}
