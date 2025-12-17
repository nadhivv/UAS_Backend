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

    achievementRoutes.Get("/", achievementService.GetAllAchievements,middleware.RequirePermission("achievement:read"))
    achievementRoutes.Get("/:id", achievementService.GetAchievementByID, middleware.RequirePermission("achievement:read"))
    achievementRoutes.Post("/", achievementService.CreateAchievement, middleware.RequirePermission("achievement:create"))
    achievementRoutes.Put("/:id", achievementService.UpdateAchievement, middleware.RequirePermission("achievement:update"))
    achievementRoutes.Delete("/:id", achievementService.DeleteAchievement, middleware.RequirePermission("achievement:delete"))
    achievementRoutes.Post("/:id/submit", achievementService.SubmitAchievement, middleware.RequirePermission("achievement:update"))
    achievementRoutes.Post("/:id/verify", middleware.RequirePermission("achievement:verify"), achievementService.VerifyAchievement)
    achievementRoutes.Post("/:id/reject", middleware.RequirePermission("achievement:verify"), achievementService.RejectAchievement)
    achievementRoutes.Get("/:id/history", achievementService.GetAchievementHistory, middleware.RequirePermission("achievement:read"))
    achievementRoutes.Post("/:id/attachments", achievementService.UploadAttachment, middleware.RequirePermission("achievement:update"))

}