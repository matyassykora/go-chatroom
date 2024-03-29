package handlers

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type User struct {
	ID       int
	Username string
	Password string
}

const Secret = "1234"

func CheckPasswordHash(hash string, password string) bool {
	ret := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return ret == nil
}

func FindByCredentials(username string, password string) (*User, error) {
	if username == "mtt" && password == "0012" {
		return &User{
			ID:       1,
			Username: username,
			Password: password,
		}, nil
	}
	if username == "pepa" && password == "abcd" {
		return &User{
			ID:       2,
			Username: username,
			Password: password,
		}, nil
	}
	return nil, errors.New("User not found")
}

func (h *hub) Login(c *fiber.Ctx) error {
	loginRequest := new(LoginRequest)
	if err := c.BodyParser(loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	user, err := FindByCredentials(loginRequest.Username, loginRequest.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	claims := jwt.MapClaims{
		"ID":       user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(Secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": err.Error(),
		})
	}

	sess, err := h.store.Get(c)
	if err != nil {
		return err
	}

	sess.Set("username", user.Username)
	// TODO: implement token stuff
	sess.Set("token", t)
	if err := sess.Save(); err != nil {
		return err
	}

	return c.Redirect("/chat", 303)
}

func (h *hub) HandleLoginGet(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return err
	}

	username := sess.Get("username")
	if username != nil {
		return c.Redirect("/chat", 303)
	}
	return c.Render("login", fiber.Map{})
}
