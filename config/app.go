package config

import (
	"time"
	"github.com/gofiber/fiber/v2"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		AppName:      "sistem pelaporan prestasi mahasiswa",
		ReadTimeout:  10 * time.Second, 
		WriteTimeout: 10 * time.Second,
		BodyLimit:    10 * 1024 * 1024, 
		JSONEncoder:  nil,              
		JSONDecoder:  nil,
	}
}