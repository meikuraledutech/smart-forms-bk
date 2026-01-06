package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTAuthMiddleware validates access token and injects user_id
func JWTAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.ErrUnauthorized
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.ErrUnauthorized
		}

		tokenString := parts[1]

		claims, err := ValidateAccessToken(tokenString)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		// 🔑 THIS is what Forms handlers rely on
		c.Locals("user_id", claims.UserID)

		// Store role (default to 'user' for backwards compatibility with old tokens)
		role := claims.Role
		if role == "" {
			role = "user"
		}
		c.Locals("user_role", role)

		return c.Next()
	}
}

// RequireSuperAdmin middleware ensures user has super_admin role
func RequireSuperAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || role != "super_admin" {
			return fiber.NewError(fiber.StatusForbidden, "Super admin access required")
		}
		return c.Next()
	}
}
