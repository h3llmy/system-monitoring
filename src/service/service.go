package service

import (
	"sync"
	"time"
)

var (
	mu         sync.Mutex
	prevTime   time.Time
	maxHistory = 60
)
