package config

import	(
	"log"

)

func LogInfo(message string) {
	log.Println("info:", message)
}

func LogError(message string) {
	log.Println("error:", message)
}