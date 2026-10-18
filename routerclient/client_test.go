package routerclient

import "testing"

func TestBuildRouteRequest_HeadingAndExclusions(t *testing.T) {
	heading, tolerance := 90, 45
	req := buildRouteRequest(RouteQuery{
		From:                 Coordinate{Latitude: 51.0, Longitude: 4.0},
		To:                   Coordinate{Latitude: 51.1, Longitude: 4.1},
		Vehicle:              "motorcycle",
		FromHeading:          &heading,
		FromHeadingTolerance: &tolerance,
		ExcludePolygons: [][]Coordinate{
			{{Latitude: 51.0, Longitude: 4.0}, {Latitude: 51.0, Longitude: 4.001}, {Latitude: 51.001, Longitude: 4.001}},
		},
	})

	first := req.Locations[0]
	if first.PreferredHeading == nil || *first.PreferredHeading != 90 {
		t.Fatalf("heading: %v", first.PreferredHeading)
	}
	if first.HeadingTolerance == nil || *first.HeadingTolerance != 45 {
		t.Fatalf("tolerance: %v", first.HeadingTolerance)
	}
	if last := req.Locations[len(req.Locations)-1]; last.PreferredHeading != nil {
		t.Fatalf("only the first location gets a heading, got %v", last.PreferredHeading)
	}
	if len(req.ExcludePolygons) != 1 || len(req.ExcludePolygons[0].Points) != 3 {
		t.Fatalf("polygons: %v", req.ExcludePolygons)
	}
	if p := req.ExcludePolygons[0].Points[1]; p.Lat != 51.0 || p.Lon != 4.001 {
		t.Fatalf("polygon point: %v", p)
	}
}

func TestBuildRouteRequest_NoHeadingByDefault(t *testing.T) {
	req := buildRouteRequest(RouteQuery{
		From: Coordinate{Latitude: 51.0, Longitude: 4.0},
		To:   Coordinate{Latitude: 51.1, Longitude: 4.1},
	})
	if req.Locations[0].PreferredHeading != nil || req.Locations[0].HeadingTolerance != nil {
		t.Fatal("heading must stay unset")
	}
	if len(req.ExcludePolygons) != 0 {
		t.Fatal("no exclusions expected")
	}
}
