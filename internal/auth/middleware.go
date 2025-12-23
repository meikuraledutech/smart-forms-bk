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

		return c.Next()
	}
}
