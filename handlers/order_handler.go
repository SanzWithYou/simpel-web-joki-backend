package handlers

import (
	"backend/dto"
	"backend/models"
	"backend/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Handler order
type OrderHandler struct {
	DB *gorm.DB
}

// Konstruktor
func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{DB: db}
}

// Buat order
func (h *OrderHandler) CreateOrderWithFile(c *fiber.Ctx) error {
	// Ambil file
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to get file", err)
	}

	// Ambil data
	username := c.FormValue("username")
	password := c.FormValue("password")
	joki := c.FormValue("joki")

	// Validasi
	if username == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username wajib diisi", nil)
	}
	if password == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Password wajib diisi", nil)
	}
	if joki == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Jenis joki wajib diisi", nil)
	}

	// Buka file
	fileData, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to open file", err)
	}
	defer fileData.Close()

	// Simpan S3
	url, err := utils.SaveUploadToS3(fileData, file, "bukti_transfer")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save file", err)
	}

	// Request
	req := dto.CreateOrderRequest{
		Username:      username,
		Password:      password,
		Joki:          joki,
		BuktiTransfer: url, // Now storing URL instead of local path
	}

	// Enkripsi user
	encryptedUsername, err := utils.EncryptWithAES(req.Username)
	if err != nil {
		_ = utils.DeleteFileFromS3(url) // Delete from S3 if encryption fails
		log.Println("Error encrypting username:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengenkripsi username", err)
	}

	// Enkripsi pass
	encryptedPassword, err := utils.EncryptWithAES(req.Password)
	if err != nil {
		_ = utils.DeleteFileFromS3(url) // Delete from S2 if encryption fails
		log.Println("Error encrypting password:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengenkripsi password", err)
	}

	// Model
	order := models.Order{
		Username:      encryptedUsername,
		Password:      encryptedPassword,
		Joki:          req.Joki,
		BuktiTransfer: req.BuktiTransfer, // Storing URL
	}

	// Simpan DB
	if err := h.DB.Create(&order).Error; err != nil {
		_ = utils.DeleteFileFromS3(url) // Delete from S3 if DB save fails
		log.Println("Error creating order:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat order", err)
	}

	// Notifikasi
	go func() {
		err := utils.SendNewOrderNotificationEmail(order.ID, username, joki, order.BuktiTransfer)
		if err != nil {
			log.Printf("Failed to send admin notification email: %v", err)
		} else {
			log.Printf("Admin notification email sent for order %d", order.ID)
		}
	}()

	// Response
	response := dto.OrderResponse{
		ID:            order.ID,
		Username:      req.Username,
		Password:      req.Password,
		Joki:          order.Joki,
		BuktiTransfer: order.BuktiTransfer, // Return URL
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Order berhasil dibuat", response)
}

// Ambil semua
func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	var orders []models.Order

	// Query DB
	if err := h.DB.Find(&orders).Error; err != nil {
		log.Println("Error getting orders:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data order", err)
	}

	// Konversi
	response := make([]dto.OrderListResponse, 0)
	for _, order := range orders {
		response = append(response, dto.OrderListResponse{
			ID:            order.ID,
			Joki:          order.Joki,
			BuktiTransfer: order.BuktiTransfer, // Return URL
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Data order berhasil diambil", response)
}

// Ambil satu
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")

	// Validasi ID
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID parameter wajib diisi", nil)
	}

	// Cari order
	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Order tidak ditemukan", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mencari order", err)
	}

	// Dekripsi user
	decryptedUsername, err := utils.DecryptWithAES(order.Username)
	if err != nil {
		log.Println("Error decrypting username:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mendekripsi username", err)
	}

	// Dekripsi pass
	decryptedPassword, err := utils.DecryptWithAES(order.Password)
	if err != nil {
		log.Println("Error decrypting password:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mendekripsi password", err)
	}

	// Response
	response := dto.OrderResponse{
		ID:            order.ID,
		Username:      decryptedUsername,
		Password:      decryptedPassword,
		Joki:          order.Joki,
		BuktiTransfer: order.BuktiTransfer, // Return URL
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Order ditemukan", response)
}

// Hapus order
func (h *OrderHandler) DeleteOrder(c *fiber.Ctx) error {
	id := c.Params("id")

	// Validasi ID
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID parameter wajib diisi", nil)
	}

	// Cari order
	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Order tidak ditemukan", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mencari order", err)
	}

	// Hapus file
	if order.BuktiTransfer != "" {
		_ = utils.DeleteFileFromS3(order.BuktiTransfer)
	}

	// Hapus DB
	if err := h.DB.Delete(&order).Error; err != nil {
		log.Println("Error deleting order:", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus order", err)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Order berhasil dihapus", nil)
}
