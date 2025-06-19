package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/monitor"
)

var (
	MonitorConfig monitor.Config = monitor.Config{
		Title:   "System Monitoring",
		Refresh: time.Duration(1000) * time.Millisecond, // Refresh interval in seconds
	}
)
