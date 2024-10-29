package receipt

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/receipt-processor/receipt/model"
)

type receiptStore interface {
	SaveReceipt(receipt *model.Receipt) (*uuid.UUID, error)
	ReadReceipt(id uuid.UUID) (*model.Receipt, error)
}
type ReceiptService struct {
	receiptStore receiptStore
}

type ReceiptServiceConfig struct {
	ReceiptStore receiptStore
}

func NewReceiptService(config *ReceiptServiceConfig) *ReceiptService {
	return &ReceiptService{
		receiptStore: config.ReceiptStore,
	}
}

func (s *ReceiptService) ProcessReceipt(receipt *model.Receipt) (*uuid.UUID, error) {
	err := validateReceipt(receipt)

	if err != nil {
		return nil, fmt.Errorf("invalid receipt data:  %s", err)
	}
	return s.receiptStore.SaveReceipt(receipt)
}

func (s *ReceiptService) GetReceipt(receiptId uuid.UUID) (*model.Receipt, error) {
	return s.receiptStore.ReadReceipt(receiptId)
}

func (s *ReceiptService) CalculatePoints(receiptId uuid.UUID) (int64, error) {
	receipt, err := s.receiptStore.ReadReceipt(receiptId)

	if err != nil {
		log.Println("error retrieving receipt")
		return -1, err
	}
	points, err := calculatePointsFromReceipt(receipt)

	if err != nil {
		log.Println("error calculating receipt")
		return -1, err
	}
	return points, nil
}
func countAlphanNumericCharacters(s string) int64 {
	var count int64 = 0
	for _, c := range s {
		if unicode.IsLetter(c) {
			count += 1
		}
	}
	return count
}
func calculatePointsFromReceipt(receipt *model.Receipt) (int64, error) {
	var points int64

	// calculate retailer number of characters
	points += countAlphanNumericCharacters(receipt.Retailer)

	// 50 points if the total is a round dollar amount with no cents.
	// this program excludes 0 as an exception since author assumes that a receipt with zero total
	// should not earn that much point
	if int64(receipt.Total) != 0 && float64(int(receipt.Total)) == receipt.Total {
		points += 50
	}
	// 25 points if the total is a multiple of 0.25
	if int64(receipt.Total) != 0 && math.Mod(receipt.Total, 0.25) == 0 {
		points += 25
	}

	// 5 points for every two items on the receipt.
	points += int64(len(receipt.Items)/2) * 5

	// multiply the price by 0.2 if description length is multiple of 3
	pointsByDescription := 0
	for _, item := range receipt.Items {
		// check if description length is multiple of 3
		descriptionLength := len(strings.TrimSpace(item.ShortDescription))
		if descriptionLength != 0 && descriptionLength%3 == 0 {
			pointsByDescription += int(math.Ceil(item.Price * 0.2))
		}
	}
	points += int64(pointsByDescription)

	// 6 points if the day in the purchase date is odd.
	parsedDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		return -1, errors.New("cannot parse date to number")
	}
	if parsedDate.Day()%2 != 0 {
		points += 6
	}
	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	receiptTime, err := time.Parse("15:04", receipt.PurchaseTime)

	if err != nil {
		return -1, errors.New("invalid time format")
	}

	startTime, _ := time.Parse("15:04", "14:00") // 2 p.m.
	endTime, _ := time.Parse("15:04", "16:00")   // 4 p.m.
	if receiptTime.After(startTime) && receiptTime.Before(endTime) {
		points += 10
	}
	return points, nil
}

func validateReceipt(receipt *model.Receipt) error {
	// parse time
	_, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		return errors.New("invalid purchase date format")
	}

	_, err = time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		return errors.New("invalid purchase time format")
	}

	if receipt.Total < 0 {
		return errors.New("invalid total price")
	}

	if len(receipt.Items) == 0 {
		return errors.New("items can not be empty")
	}
	return nil
}
