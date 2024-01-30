package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	// "github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"

	jwtware "github.com/gofiber/contrib/jwt"

	"github.com/matyassykora/go-chatroom/internal/handlers"
	"github.com/matyassykora/go-chatroom/internal/middleware"
)

func NewAuthMiddleware(secret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
	})
}

func main() {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:        engine,
		ViewsLayout:  "layouts/base",
		ErrorHandler: handlers.HandleErrors,
	})

	// TODO: disable
	engine.Reload(true)

	app.Static("/", "./public")

	// openssl rand -base64 32
	// app.Use(encryptcookie.New(encryptcookie.Config{
	// 	Key: "IOa1yLkd7RaD6YWUUCGuApTgIn1uqq4rqPTE227Stck=",
	// }))
	app.Use(middleware.Compress)
	app.Use(middleware.Helmet)
	// app.Use(middleware.BasicAuth)
	app.Use(middleware.Limiter)

	app.Get("/metrics", middleware.Monitor)

	store := session.New()
	hub := handlers.NewHub(store)
	go hub.Run()

	// jwt := NewAuthMiddleware(handlers.Secret)

	app.Post("/chat", hub.Login)
	app.Get("/chat", hub.HandleChatsGet)
	app.Get("/", hub.HandleLoginGet)

	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return err
		}

		sess.Destroy()
		return c.Redirect("/")
	})

	// app.Get("/chat", jwt, handlers.HandleProtectedRoute, handlers.HandleChatsGet)

	// app.Get("/protected", jwt, handlers.HandleProtectedRoute, func(c *fiber.Ctx) error {return c.SendString("PEPA")})
	// app.Get("/", handlers.Login, jwt, handlers.HandleProtectedRoute, handlers.HandleChatsGet)

	app.Use("/ws", hub.HandleWebsocketUpgrade)
	app.Get("/ws", websocket.New(hub.HandleWebsockets))

	log.Fatal(app.Listen(":3000"))
}
