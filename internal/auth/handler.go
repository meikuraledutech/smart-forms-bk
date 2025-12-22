package auth

import "github.com/gofiber/fiber/v2"

// AuthHandler holds dependencies for HTTP layer
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates handler
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

/*
========================
 LOGIN
========================
*/

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	access, refresh, err := h.service.Login(
		c.Context(),
		req.Username,
		req.Password,
	)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	return c.JSON(loginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

/*
========================
 REFRESH
========================
*/

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	access, err := h.service.RefreshAccessToken(
		c.Context(),
		req.RefreshToken,
	)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	return c.JSON(refreshResponse{
		AccessToken: access,
	})
}

/*
========================
 REGISTER
========================
*/

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "username and password are required",
		})
	}

	err := h.service.Register(
		c.Context(),
		req.Username,
		req.Password,
	)

	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "user already registered",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "could not register user",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "user registered successfully",
	})
}

