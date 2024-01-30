package handlers

import (
	"log"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func HandleIndexGet(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{})
}

func (h *hub) HandleChatsGet(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return err
	}

	username := sess.Get("username")
	if username == nil {
		return c.Redirect("/")
	}

	log.Printf("User '%s' logged in", username)

	return c.Render("chat", fiber.Map{
		"Title":    "Chat",
		"Username": username,
	})
}

func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(Secret)},
		ErrorHandler: jwtError,
	})
	// user := c.Locals("user").(*jwt.Token)
	// claims := user.Claims.(jwt.MapClaims)
	// username := claims["username"].(string)
	// c.Locals("username", username)
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
}
