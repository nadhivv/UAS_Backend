package route

import (
	"UAS/middleware"
	"UAS/app/repository"
	"UAS/app/service"
	
	"github.com/gofiber/fiber/v2"
)

func setupAuthRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) {
	authService := service.NewAuthService(userRepo, roleRepo)
	
	authRoutes := router.Group("/auth")
	
	authRoutes.Post("/login", authService.Login)
	authRoutes.Post("/refresh", authService.RefreshToken)
	authRoutes.Post("/logout", authService.Logout)
	
	authRoutes.Get("/profile", middleware.RequireAuth(userRepo),authService.Profile,)
	authRoutes.Post("/change-password",middleware.RequireAuth(userRepo),authService.ChangePassword,)
}