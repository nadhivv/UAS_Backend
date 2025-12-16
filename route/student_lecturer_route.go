package route

import (
	"UAS/app/repository"
	"UAS/app/service"
	"UAS/database"
	"UAS/middleware"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupStudentLecturerRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	mongoDB *mongo.Database,
) {

	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PgDB)
	achievementRepo := repository.NewAchievementRepository(
		mongoDB.Collection("achievements"),
	)

	studentLecturerService := service.NewStudentLecturerService(
		studentRepo,
		lecturerRepo,
		userRepo,
		roleRepo,
		achievementRepo,
		achievementRefRepo,
	)

	students := router.Group("/students")
	students.Use(middleware.RequireAuth(userRepo))

	students.Get("/", studentLecturerService.GetAllStudents)
	students.Get("/:id", studentLecturerService.GetStudentByID)
	students.Get("/:id/achievements", studentLecturerService.GetStudentAchievements)
	students.Put("/:id/advisor", studentLecturerService.UpdateStudentAdvisor)

	lecturers := router.Group("/lecturers")
	lecturers.Use(middleware.RequireAuth(userRepo))

	lecturers.Get("/", studentLecturerService.GetAllLecturers)
	lecturers.Get("/:id/advisees", studentLecturerService.GetLecturerAdvisees)
}
