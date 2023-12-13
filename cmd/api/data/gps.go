package data


type GpsData struct {
	Timestamp string

	Latitude float64
	Longitude float64
	Altitude float64

	Satellite uint64

	Hdop float64
	Vdop float64
}