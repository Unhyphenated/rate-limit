package bucket

import (
	"time"
)

const RATE = 1
const MAX_TOKENS = 100

type Bucket struct {
	tokens int
	lastRefill time.Time
}

