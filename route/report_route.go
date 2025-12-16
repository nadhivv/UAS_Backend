package route

import (
	"UAS/app/repository"
	"UAS/app/service"
	"UAS/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupReportRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	roleRepo repository.RoleRepository,
	reportRepo repository.ReportRepository,
) {
	reportService := service.NewReportService(
		reportRepo,
		userRepo,
		studentRepo,
		lecturerRepo,
		roleRepo,
	)

	router.Get(
		"/reports/statistics",
		middleware.RequireAuth(userRepo),
		reportService.GetStatistics,
	)

	router.Get(
		"/reports/student/:id",
		middleware.RequireAuth(userRepo),
		reportService.GetStudentReport,
	)
}
