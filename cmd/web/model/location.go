package model

import "strconv"

// Location type for saving location in mongo db
type Location struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

// NewLocationFromLongLatStr helper function
func NewLocationFromLongLatStr(longitude, latitude string) Location {
	lat, _ := strconv.ParseFloat(latitude, 64)
	long, _ := strconv.ParseFloat(longitude, 64)
	return Location{
		Type: "Point",
		// In mongo db first is longitude then latitude
		Coordinates: []float64{long, lat},
	}
}

// NewLocationFromLongLat helper function
func NewLocationFromLongLat(longitude, latitude float64) Location {
	return Location{
		Type: "Point",
		// In mongo db first is longitude then latitude
		Coordinates: []float64{longitude, latitude},
	}
}

// GetLatitude get latitude
func (l Location) GetLatitude() float64 {
	if len(l.Coordinates) == 2 {
		return l.Coordinates[1]
	}
	return 0
}

// GetLongitude get longitude
func (l Location) GetLongitude() float64 {
	if len(l.Coordinates) == 2 {
		return l.Coordinates[0]
	}
	return 0
}
