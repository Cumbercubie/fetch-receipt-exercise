package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/receipt-processor/receipt"
	"github.com/receipt-processor/receipt/db"
)

// var env = os.Getenv("ENV")

func initGinServer() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus((http.StatusOK))
			return
		}

		c.Next()
	})
	receiptStore := db.InitDB()
	receiptService := receipt.NewReceiptService(&receipt.ReceiptServiceConfig{
		ReceiptStore: receiptStore,
	})
	receiptHandler := receipt.NewRouteHandler(&receipt.RouteHandlerConfig{
		ReceiptService: *receiptService,
	})
	group := r.Group("/api/v1")

	receiptHandler.RegisterReceiptRoutes(group)

	r.Run(":3000")

}

func main() {
	// ctx := context.Background()
	initGinServer()
}
