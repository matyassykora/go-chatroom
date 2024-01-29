package handlers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
)

type chat struct {
	ID      uuid.UUID
	Message string
}

// hub maintains the set of active clients and broadcasts messages to the clients
type hub struct {
	// Store of app sessions
	store *session.Store

	// Registered clients.
	clients map[*websocket.Conn]*client

	// Register requests from the clients.
	Register chan *websocket.Conn

	// Inbound messages from the clients.
	Broadcast chan string

	// Unregister requests from clients.
	Unregister chan *websocket.Conn

	// Mutex for clients
	mutex sync.Mutex
}

type message struct {
	Message string     `json:"msg"`
	Headers HtmxHeader `json:"HEADERS"`
}

type HtmxHeader struct {
	HxRequest     string  `json:"Hx-Request"`
	HxTrigger     string  `json:"Hx-Trigger"`
	HxTriggerName *string `json:"Hx-Trigger-Name"`
	HxTarget      string  `json:"Hx-Target"`
	HxCurrentURL  string  `json:"Hx-Current-URL"`
}

// client is a middleman between the websocket connection and the hub.
type client struct {
	Username string
	// active bool
}

type server struct {
	messageChan  chan string
	pastMessages []string
	mutex        sync.Mutex
	hub          *hub
}

func (h *hub) Run() {
	for {
		select {
		case connection := <-h.Register:
			username := connection.Locals("username").(string)
			h.clients[connection] = &client{
				Username: username,
			}
			log.Println("connection registered")

		case msg := <-h.Broadcast:
			// log.Println("message received:", msg)

			// Send the message to all clients
			for connection := range h.clients {
				if err := connection.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					log.Println("write error:", err)

					h.Unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-h.Unregister:
			// TODO: fix
			// logoutMessage := h.clients[connection].Username + " logged out"
			// connection.WriteMessage(websocket.TextMessage, []byte(logoutMessage))

			// Remove the client from the hub
			delete(h.clients, connection)

			log.Println("connection unregistered")
		}
	}
}

// NewHub creates new Hub
func NewHub(s *session.Store) *hub {
	return &hub{
		store:      s,
		clients:    make(map[*websocket.Conn]*client),
		Register:   make(chan *websocket.Conn),
		Broadcast:  make(chan string),
		Unregister: make(chan *websocket.Conn),
	}
}

func (h *hub) HandleWebsocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		sess, err := h.store.Get(c)
		if err != nil {
			return err
		}
		username := sess.Get("username")
		if username == nil {
			return fiber.ErrInternalServerError
		}
		c.Locals("username", username)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (h *hub) HandleWebsockets(c *websocket.Conn) {
	// c.Locals is added to the *websocket.Conn
	// log.Println(c.Locals("allowed"))  // true
	// log.Println(c.Params("id"))       // 123
	// log.Println(c.Query("v"))         // 1.0
	// log.Println(c.Cookies("session")) // ""

	username := c.Locals("username").(string)

	log.Printf("websocket user: %s", username)

	defer func() {
		h.Unregister <- c
		c.Close()
	}()

	h.Register <- c

	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			return
		}

		if messageType == websocket.TextMessage {
			var v message
			err := json.Unmarshal(msg, &v)
			if err != nil {
				log.Println(err)
				h.Broadcast <- "ERROR WHEN SENDING MESSAGE"
				return
			}
			if v.Message == "" {
				return
			}
			h.Broadcast <- "<div hx-swap-oob='beforeend:#log'><p>" + username + ": " + v.Message + "</p></div>"
		} else {
			log.Println("message type:", messageType)
		}
	}
}
