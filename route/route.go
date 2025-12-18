package route

import (
	"UAS/database"
	"UAS/app/repository"
	"UAS/app/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetupRoutes(app *fiber.App) {
	db := database.PgDB

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	reportRepo := repository.NewReportRepository()

	userService := service.NewUserService(userRepo, roleRepo, studentRepo, lecturerRepo)

	examAPI := app.Group("/uas/api")

	setupAuthRoutes(examAPI, userRepo, roleRepo, studentRepo, lecturerRepo)
	setupUserRoutes(examAPI, userService, userRepo, roleRepo)

	SetupReportRoutes(
		examAPI,
		userRepo,
		studentRepo,
		lecturerRepo,
		roleRepo,
		reportRepo,
	)

	examAPI.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "cobacoba",
			"version": "1.0",
		})
	})

	examAPI.Get("/swagger/*", swagger.HandlerDefault)
}
