package handlers

import (
	"CraftTanks/database"
	"CraftTanks/models"
	"CraftTanks/utils"
	"os"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

const (
	DefaultAccessTokenExp  = 900
	DefaultRefreshTokenExp = 604800
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GenerateTokens(username string) (string, string, error) {

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("Error loading .env file")
	}

	accessExp := utils.GetEnvAsInt("ACCESS_TOKEN_EXPIRES", DefaultAccessTokenExp)
	refreshExp := utils.GetEnvAsInt("REFRESH_TOKEN_EXPIRES", DefaultRefreshTokenExp)

	accessClaims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Second * time.Duration(accessExp)).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Second * time.Duration(refreshExp)).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	err = database.RedisClient.Set(database.Ctx, username, refreshTokenString, time.Second*time.Duration(refreshExp)).Err()
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func Register(c *fiber.Ctx) error {
	log.Println("Registering...")

	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Hash the password before saving
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not hash password"})
	}
	user.Password = hashedPassword

	database.DB.Create(&user)
	return c.JSON(fiber.Map{"message": "User registered successfully"})
}

func Login(c *fiber.Ctx) error {
	var input models.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user models.User
	database.DB.Where("username = ?", input.Username).First(&user)
	if user.ID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if !utils.CheckPassword(user.Password, input.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	accessToken, refreshToken, err := GenerateTokens(user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token generation failed"})
	}

	err = database.SetSession(user.Username, refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not store session"})
	}

	database.TrackActiveUser(user.Username)

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func Logout(c *fiber.Ctx) error {
	userID := c.Locals("username").(string)

	err := database.DeleteSession(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not remove session"})
	}

	database.RemoveActiveUser(userID)

	return c.JSON(fiber.Map{"message": "Logout successful"})
}

func SessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No token provided"})
		}

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		username := claims["username"].(string)

		storedToken, err := database.GetSession(username)
		if err != nil || storedToken != token {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Session expired or invalid"})
		}

		c.Locals("username", username)

		return c.Next()
	}
}

func GetActiveUsers(c *fiber.Ctx) error {
	users, err := database.GetActiveUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve active users"})
	}
	return c.JSON(fiber.Map{"active_users": users})
}

func Refresh(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var request RefreshRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	token, err := jwt.Parse(request.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	username := claims["username"].(string)

	// Verify refresh token from Redis
	storedToken, err := database.RedisClient.Get(database.Ctx, username).Result()
	if err != nil || storedToken != request.RefreshToken {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
	}

	// Generate new access and refresh tokens
	accessToken, newRefreshToken, err := GenerateTokens(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token generation failed"})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func GetUsers(c *fiber.Ctx) error {
	log.Println("Handling GET /api/users request...")

	var users []models.User
	database.DB.Select("id, username").Find(&users)

	return c.JSON(users)
}
