package profile

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
	profileProto "profile_service/internal/services/annotation_proto"
	"strconv"
)

type SuccessResponse struct {
	Data         interface{} `json:"data"`
	TotalRecords int         `json:"totalRecords"`
	Limit        int         `json:"limit"`
	Offset       int         `json:"offset"`
}

type UpdateProfileRequest struct {
	UserID string `json:"user_id"` // UUID пользователя
	Name   string `json:"name"`    // Новое имя пользователя
	Email  string `json:"email"`   // Новый email
	Phone  string `json:"phone"`   // Новый телефон
}

func (s *ProfileServiceServer) GetProfileByTokenHTTP(c *gin.Context) {
	// Извлекаем токен из запроса
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Создаем gRPC-запрос
	req := &profileProto.GetProfileByTokenRequest{Token: token}

	// Вызываем gRPC-метод
	resp, err := s.GetProfileByToken(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат в формате JSON
	c.JSON(http.StatusOK, gin.H{
		"profile": resp.Profile,
	})
}

func (s *ProfileServiceServer) GetProfileByUserIDHTTP(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	req := &profileProto.GetProfileByUserIDRequest{UserId: userID}

	resp, err := s.GetProfileByUserID(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": resp.Profile,
	})
}

func (s *ProfileServiceServer) GetProfileByIDHTTP(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	req := &profileProto.GetProfileRequest{Id: id}

	resp, err := s.GetProfile(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": resp.Profile,
	})
}

func (s *ProfileServiceServer) CreateProfileHTTP(c *gin.Context) {

	var req profileProto.CreateProfileRequest

	// Чтение тела запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Восстановление тела для возможного дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Десериализация JSON в Protobuf
	if err := protojson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//fmt.Printf("Parsed Request: %+v\n", req)

	resp, err := s.CreateProfile(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"profile_id": resp.Id,
	})
}

func (s *ProfileServiceServer) GetProfilesHTTP(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	req := &profileProto.GetProfilesRequest{Limit: limitStr, Offset: offsetStr}

	resp, err := s.GetProfiles(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении профилей"})
		return
	}

	// Преобразуем gRPC-профили в плоский формат для HTTP-ответа
	flatProfiles := make([]map[string]interface{}, len(resp.Profiles))
	for i, profile := range resp.Profiles {
		flatProfiles[i] = map[string]interface{}{
			"id":          profile.Id,
			"user_id":     profile.UserId,
			"first_name":  profile.FirstName,
			"last_name":   profile.LastName,
			"middle_name": profile.MiddleName,
			"phone":       profile.Phone,
			"address":     profile.Address,
			"birthday":    profile.Birthday,
			"created_at":  profile.CreatedAt,
			"updated_at":  profile.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:         flatProfiles,
		TotalRecords: int(resp.TotalRecords),
		Limit:        limit,
		Offset:       offset,
	})
}

func (s *ProfileServiceServer) UpdateProfileHTTP(c *gin.Context) {
	var req profileProto.UpdateProfileRequest

	// Извлечение user_id из параметров пути
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Чтение тела запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Восстановление тела для возможного дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Десериализация JSON в Protobuf
	if err := protojson.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserId = userID

	// Вызов gRPC-метода
	resp, err := s.UpdateProfile(c, &req) // c.Request.Context()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (s *ProfileServiceServer) DeleteProfileHTTP(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	req := &profileProto.DeleteProfileByIDRequest{Id: id}

	resp, err := s.DeleteProfileByID(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//c.JSON(http.StatusOK, gin.H{"message": "Profile deleted"})
	c.JSON(http.StatusOK, gin.H{"message": resp.Message})
}
