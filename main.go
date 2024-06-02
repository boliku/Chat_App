package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrader para WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Mapa de sesiones
var sessions = make(map[string][]*websocket.Conn)
var mutex = &sync.Mutex{}

// Estructura para respuesta del string aleatorio
type RandomStringResponse struct {
	SessionID string `json:"session_id"`
}

// Generar string aleatorio
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Middleware para habilitar CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()

	// Aplicar el middleware CORS
	r.Use(CORSMiddleware())

	// Endpoint para obtener el string aleatorio
	r.GET("/random-string", func(c *gin.Context) {
		sessionID := generateRandomString(10)
		c.JSON(http.StatusOK, RandomStringResponse{SessionID: sessionID})
	})

	// Endpoint para WebSocket
	r.GET("/ws/:sessionID", func(c *gin.Context) {
		sessionID := c.Param("sessionID")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer conn.Close()

		mutex.Lock()
		sessions[sessionID] = append(sessions[sessionID], conn)
		mutex.Unlock()

		for {
			// Leer mensaje
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// Enviar mensaje a todas las conexiones en la sesión
			mutex.Lock()
			for _, c := range sessions[sessionID] {
				if c != conn {
					if err := c.WriteMessage(messageType, message); err != nil {
						break
					}
				}
			}
			mutex.Unlock()
		}

		// Remover conexión de la sesión al desconectar
		mutex.Lock()
		for i, c := range sessions[sessionID] {
			if c == conn {
				sessions[sessionID] = append(sessions[sessionID][:i], sessions[sessionID][i+1:]...)
				break
			}
		}
		mutex.Unlock()
	})

	// Iniciar servidor en puerto 8080
	r.Run(":8080")
}