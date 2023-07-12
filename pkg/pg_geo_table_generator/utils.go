package pg_geo_table_generator

import (
	"math/rand"
	"time"
)

func generateRandomIntNumberOnTheSeg(begin, end int) int {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	return begin + generator.Intn(end)
}

func generateRandomFloatNumberOnTheSeg(begin, end float64) float64 {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	return begin + generator.Float64()*(end-begin)
}
