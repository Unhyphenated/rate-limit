package models


const RATE = 1
const MAX_TOKENS = 100

type Bucket struct {
	Tokens int64
	LastRefill int64
}

