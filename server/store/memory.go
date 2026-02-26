package store

import "sync"

// sync.Map: 동시 요청에 안전한 인메모리 저장소 (나중에 DB로 교체)
var AuthCodes     sync.Map
var RefreshTokens sync.Map
