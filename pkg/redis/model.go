package redis

import "time"

const UserLocationTTL = time.Minute * 10

type UserLocation struct {
	Latitude  float32
	Longitude float32
}
