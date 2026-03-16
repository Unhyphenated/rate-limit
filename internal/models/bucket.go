package models


const RATE = 1
const MAX_TOKENS = 100

type Bucket struct {
    Tokens     int64 `redis:"tokens"`
    LastRefill int64 `redis:"last_refill"`
}
