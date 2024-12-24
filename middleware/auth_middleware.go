package middleware

import (
	"context"
	"net/http"
	authpb "profile_service/internal/services/auth_proto"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Токен должен быть в формате "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Отправляем токен в Auth-сервис для проверки
		resp, err := authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil || !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": resp.GetError()})
			c.Abort()
			return
		}

		// Сохраняем user_id и role в контексте для использования в обработчиках
		c.Set("user_id", resp.GetUserId())
		c.Set("role", resp.GetRole())

		// Передаем управление следующему обработчику
		c.Next()
	}
}
