package main

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var db *gorm.DB
var jwtSecret = []byte("tu_secreto")

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func main() {
	// Configurar la conexión a la base de datos
	dsn := "host=localhost user=postgres password=Kamezennin dbname=chat_app port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("No se pudo conectar a la base de datos:", err)
		return
	}

	// Migrar modelos
	db.AutoMigrate(&User{})

	r := gin.Default()

	// Endpoints
	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.GET("/protected", authMiddleware(), protectedHandler)

	// Iniciar el servidor
	r.Run(":8080")
}

func protectedHandler(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Hello, %s!", user.(*jwt.Token).Claims.(*Claims).Username)})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		// Añadir log para verificar el token recibido
		fmt.Println("Received token:", tokenString)

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			fmt.Println("Error parsing token:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired authorization token"})
			c.Abort()
			return
		}
		if !token.Valid {
			fmt.Println("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired authorization token"})
			c.Abort()
			return
		}

		c.Set("user", token)
		c.Next()
	}
}

func registerHandler(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verificar si el nombre de usuario ya existe
	var existingUser User
	result := db.Where("username = ?", newUser.Username).First(&existingUser)
	if result.RowsAffected != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	newUser.Password = string(hashedPassword)

	// Crear usuario en la base de datos
	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func loginHandler(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Buscar usuario en la base de datos
	var user User
	result := db.Where("username = ?", loginData.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Verificar la contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generar token JWT
	token, err := generateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "User Login Successful"})
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func generateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expira en 24 horas
			IssuedAt:  time.Now().Unix(),
			Issuer:    "chat_app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}