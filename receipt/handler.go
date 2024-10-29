package receipt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/receipt-processor/receipt/model"
)

type RouteHandler struct {
	receiptService ReceiptService
}

type RouteHandlerConfig struct {
	ReceiptService ReceiptService
}

func NewRouteHandler(cfg *RouteHandlerConfig) *RouteHandler {
	return &RouteHandler{
		receiptService: cfg.ReceiptService,
	}
}
func (rh *RouteHandler) RegisterReceiptRoutes(r *gin.RouterGroup) {
	receiptGroup := r.Group("/receipts")

	receiptGroup.POST("/process", rh.ProcessReceipt)
	receiptGroup.GET("/:id/points", rh.GetReceiptPoints)
}

func (rh *RouteHandler) ProcessReceipt(c *gin.Context) {
	var receiptInput *model.Receipt
	if err := c.BindJSON(&receiptInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	fmt.Println(receiptInput.PurchaseDate, receiptInput.PurchaseTime)

	uuid, err := rh.receiptService.ProcessReceipt(receiptInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "error processing receipt",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"id": uuid,
	})
}

func (rh *RouteHandler) GetReceiptPoints(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println("Unable to parse id to uuid")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to parse id to uuid"})
		return
	}

	points, err := rh.receiptService.CalculatePoints(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Quotes retrieved successfully",
		"points":  points,
	})
}
