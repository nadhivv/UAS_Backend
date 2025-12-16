package route

import (
	"UAS/app/repository"
	"UAS/app/service"
	"UAS/middleware"
	"UAS/database"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupAchievementRoutes(router fiber.Router, 
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	mongoDB *mongo.Database) {

	// Inisialisasi repositories
	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PgDB)
	achievementRepo := repository.NewAchievementRepository(mongoDB.Collection("achievements"))
	
	// Inisialisasi service
	achievementService := service.NewAchievementService(
		achievementRepo,
		achievementRefRepo,
		studentRepo,
		lecturerRepo,
		userRepo,
		roleRepo,
	)

	achievementRoutes := router.Group("/achievements")
	achievementRoutes.Use(middleware.RequireAuth(userRepo))

	achievementRoutes.Get("/", achievementService.GetAllAchievements)
	achievementRoutes.Get("/:id", achievementService.GetAchievementByID)
	achievementRoutes.Post("/", achievementService.CreateAchievement)
	achievementRoutes.Put("/:id", achievementService.UpdateAchievement)
	achievementRoutes.Delete("/:id", achievementService.DeleteAchievement)
	achievementRoutes.Post("/:id/submit", achievementService.SubmitAchievement)
	achievementRoutes.Post("/:id/verify", achievementService.VerifyAchievement)
	achievementRoutes.Post("/:id/reject", achievementService.RejectAchievement)
	
	// achievementRoutes.Get("/:id/history", achievementService.GetAchievementHistory)
	
	// achievementRoutes.Post("/:id/attachments", achievementService.UploadAttachment)
}