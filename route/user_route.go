package route

import (
	"UAS/app/repository"
	"UAS/app/service"
	"UAS/middleware"

	"github.com/gofiber/fiber/v2"
)

func setupUserRoutes(
	router fiber.Router, 
	userService *service.UserService,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) {
	userRoutes := router.Group("/users")
	
	// ALL ROLES
	user := userRoutes.Group("",)
	userRoutes.Get("/", userService.GetAll)
	userRoutes.Get("/:id", userService.GetByID)
	userRoutes.Get("/search", userService.SearchByName)

	// ADMIN ONLY ROUTES
	user.Post("/", userService.Create, middleware.RequireAuth(userRepo),middleware.AdminOnly(roleRepo))
	user.Put("/:id", userService.Update, middleware.RequireAuth(userRepo),middleware.AdminOnly(roleRepo))
	user.Delete("/:id", userService.Delete, middleware.RequireAuth(userRepo),middleware.AdminOnly(roleRepo))
	user.Put("/:id/role", userService.UpdateRole, middleware.RequireAuth(userRepo),middleware.AdminOnly(roleRepo))
	user.Get("/inactive", userService.GetInactiveUsers, middleware.RequireAuth(userRepo),middleware.AdminOnly(roleRepo))
}