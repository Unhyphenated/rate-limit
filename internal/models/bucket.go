package models


type Bucket struct {
    Tokens     int64 `redis:"tokens"`
    LastRefill int64 `redis:"last_refill"`
}
