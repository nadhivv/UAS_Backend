package config

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

func LoggerConfig() logger.Config {
	// Pilih format berdasarkan environment
	var format string
	if os.Getenv("APP_ENV") == "production" {
		// JSON format untuk production (lebih mudah di-parse oleh log aggregator)
		format = `{"time":"${time}", "status":${status}, "method":"${method}", "path":"${path}", "latency":"${latency}", "ip":"${ip}", "user_agent":"${ua}", "error":"${error}"}` + "\n"
	} else {
		// Human readable format untuk development
		format = "[${time}] ${status} - ${method} ${path} (${latency}) | IP: ${ip} | UA: ${ua}\n"
		
		// Alternatif format yang lebih sederhana
		// format = "[${time}] ${status} - ${method} ${path} (${latency})\n"
	}
	
	return logger.Config{
		Format:       format,
		TimeFormat:   "2006-01-02 15:04:05",
		TimeZone:     "Asia/Jakarta",
		TimeInterval: 500 * time.Millisecond,
		Output:       getLoggerOutput(), // Bisa ke stdout atau file
	}
}

func getLoggerOutput() *os.File {
	// Jika ingin log ke file, bisa diatur di environment
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// Jika gagal buka file, fallback ke stdout
			return os.Stdout
		}
		return file
	}
	return os.Stdout
}