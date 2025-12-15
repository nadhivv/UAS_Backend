package route

import (
    "UAS/database"
    "UAS/app/repository"
    "UAS/app/service"

    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    db := database.PgDB
    
    userRepo := repository.NewUserRepository(db)
    roleRepo := repository.NewRoleRepository(db)
    studentRepo := repository.NewStudentRepository(db)
    lecturerRepo := repository.NewLecturerRepository(db)
    
    userService := service.NewUserService(userRepo, roleRepo, studentRepo, lecturerRepo)
    
    examAPI := app.Group("/uas/api")
    
    setupAuthRoutes(examAPI, userRepo, roleRepo)
    setupUserRoutes(examAPI, userService, userRepo, roleRepo)
    
    examAPI.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":  "OK",
            "service": "cobacoba",
            "version": "1.0",
        })
    })
}