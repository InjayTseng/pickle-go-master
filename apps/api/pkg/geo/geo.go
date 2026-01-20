package geo

import (
	"fmt"
	"math"
)

const (
	// EarthRadiusKm is the Earth's radius in kilometers
	EarthRadiusKm = 6371.0
	// EarthRadiusM is the Earth's radius in meters
	EarthRadiusM = 6371000.0
)

// Point represents a geographic point
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// NewPoint creates a new Point
func NewPoint(lat, lng float64) Point {
	return Point{Lat: lat, Lng: lng}
}

// Distance calculates the Haversine distance between two points in meters
func Distance(p1, p2 Point) float64 {
	lat1Rad := toRadians(p1.Lat)
	lat2Rad := toRadians(p2.Lat)
	deltaLat := toRadians(p2.Lat - p1.Lat)
	deltaLng := toRadians(p2.Lng - p1.Lng)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusM * c
}

// DistanceKm calculates the Haversine distance between two points in kilometers
func DistanceKm(p1, p2 Point) float64 {
	return Distance(p1, p2) / 1000.0
}

// IsWithinRadius checks if a point is within a given radius (in meters) of another point
func IsWithinRadius(center, point Point, radiusM float64) bool {
	return Distance(center, point) <= radiusM
}

// BoundingBox calculates a bounding box around a point with a given radius
// Returns (minLat, maxLat, minLng, maxLng)
func BoundingBox(center Point, radiusM float64) (float64, float64, float64, float64) {
	// Approximate degrees per meter
	latDelta := radiusM / 111320.0 // approximately 111.32 km per degree of latitude
	lngDelta := radiusM / (111320.0 * math.Cos(toRadians(center.Lat)))

	minLat := center.Lat - latDelta
	maxLat := center.Lat + latDelta
	minLng := center.Lng - lngDelta
	maxLng := center.Lng + lngDelta

	return minLat, maxLat, minLng, maxLng
}

// toRadians converts degrees to radians
func toRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// toDegrees converts radians to degrees
func toDegrees(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// FormatDistance formats a distance in meters to a human-readable string
func FormatDistance(meters float64) string {
	if meters < 1000 {
		return formatFloat(meters, 0) + " m"
	}
	return formatFloat(meters/1000, 1) + " km"
}

func formatFloat(f float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return fmt.Sprintf(format, f)
}
