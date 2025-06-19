package config

import "github.com/gofiber/fiber/v2/middleware/healthcheck"

var (
	HealthCheckConfig healthcheck.Config = healthcheck.Config{
		LivenessEndpoint: "/live",
	}
)
