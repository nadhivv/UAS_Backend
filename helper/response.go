package helper

import (
    "github.com/gofiber/fiber/v2"
)

func ResponseJSON(c *fiber.Ctx, status int, message string, success bool, data interface{}) error {
    return c.Status(status).JSON(fiber.Map{
        "message": message,
        "success": success,
        "data":  data,
	     })
    }