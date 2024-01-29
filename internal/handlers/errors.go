package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func HandleErrors(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	err = c.Status(code).Render("error", fiber.Map{
		"message": err.Error(),
		"code":    code,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return nil
}
