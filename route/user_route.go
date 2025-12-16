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
	userRoutes := router.Group("/users",middleware.RequireAuth(userRepo))
	
	user := userRoutes.Group("",middleware.RequireAuth(userRepo), middleware.AdminOnly(roleRepo))
	userRoutes.Get("/", userService.GetAll)
	userRoutes.Get("/:id", userService.GetByID)
	userRoutes.Get("/search", userService.SearchByName)

	user.Post("/", userService.Create)
	user.Put("/:id", userService.Update)
	user.Delete("/:id", userService.Delete)
	user.Put("/:id/role", userService.UpdateRole)
}