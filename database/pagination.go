package database

import "math"

func GetPagesNeeded(recordsNumber, pageSize int) int {
	return int(math.Ceil(float64(recordsNumber) / 200))
}
