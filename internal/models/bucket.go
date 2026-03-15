package models

import (
	"time"
)

const RATE = 1
const MAX_TOKENS = 100

type Bucket struct {
	Tokens int
	LastRefill time.Time
}

