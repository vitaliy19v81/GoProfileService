package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"profile_service/config"
	"profile_service/database"
	authpb "profile_service/internal/services/auth_proto"
	"profile_service/internal/services/proto"
	"profile_service/middleware"
	ps "profile_service/profile_service"
	rProfile "profile_service/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию (переменные окружения)
	config.LoadConfig()

	// Подключаемся к Auth-сервису
	authClient := createAuthClient("localhost:50051")

	// Подключение к базе данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Выполнение миграций
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		// Укажите домен, с которого разрешены запросы (например, ваш фронтенд)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true") // Для авторизации через cookie
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Подключаем middleware
	r.Use(middleware.AuthMiddleware(authClient))

	//profileService := rProfile.NewProfileService(db)

	profilesGroup := r.Group("/api/profiles")
	{
		profilesGroup.POST("/create", rProfile.CreateProfile(db))
		profilesGroup.GET("/:id", rProfile.GetProfile(db))
	}

	// gRPC маршруты
	go func() {

		grpcServer := grpc.NewServer()

		profileService := ps.NewProfileService(db)                     // Создание сервиса для работы с профилями
		proto.RegisterProfileServiceServer(grpcServer, profileService) // Регистрируем ProfileServiceServer для gRPC

		// Включение reflection для gRPC
		reflection.Register(grpcServer)

		listener, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen on port 50052: %v", err)
		}

		log.Println("gRPC server is running on port 50052")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}

	}()

	go func() {
		if err := r.Run("127.0.0.1:8082"); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	//log.Println("HTTP server started on :8080")
	//if err := r.Run(":8080"); err != nil {
	//	log.Fatalf("HTTP server failed: %v", err)
	//}

	select {} // Программа не завершится и будет продолжать работать
}

func createAuthClient(address string) authpb.AuthServiceClient {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к Auth-сервису: %v", err)
	}
	return authpb.NewAuthServiceClient(conn)
}
