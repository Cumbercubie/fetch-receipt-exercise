package receipt

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/receipt-processor/receipt/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReceiptStore struct {
	mock.Mock
}

var testReceipt *model.Receipt = &model.Receipt{
	Retailer:     "TestRetailer",
	PurchaseDate: "1999-01-20",
	PurchaseTime: "13:22",
	Items:        []model.Item{},
	Total:        12.12,
}

func (m *MockReceiptStore) ReadReceipt(id uuid.UUID) (*model.Receipt, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Receipt), nil
}

func (m *MockReceiptStore) SaveReceipt(receipt *model.Receipt) (*uuid.UUID, error) {
	args := m.Called(receipt)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*uuid.UUID), nil
}

func TestProcessReceiptService(t *testing.T) {
	testReceipt := &model.Receipt{
		Retailer:     "TestRetailer",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:22",
		Items: []model.Item{
			{
				ShortDescription: "test product description",
				Price:            0,
			},
		},
		Total: 0,
	}

	testUuid := uuid.New()

	mockReceiptStore := &MockReceiptStore{}
	receiptService := &ReceiptService{
		receiptStore: mockReceiptStore,
	}
	mockReceiptStore.On("SaveReceipt", mock.Anything).Return(&testUuid, nil)

	actualUuid, err := receiptService.ProcessReceipt(testReceipt)

	assert.Nil(t, err)

	assert.Equal(t, testUuid, *actualUuid)
}

func TestProcessReceiptServiceInvalidReceipt(t *testing.T) {
	testReceipt := &model.Receipt{
		Retailer:     "TestRetailer",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:22",
		Items:        []model.Item{},
		Total:        0,
	}

	mockReceiptStore := &MockReceiptStore{}
	receiptService := &ReceiptService{
		receiptStore: mockReceiptStore,
	}

	testUuid := uuid.New()

	mockReceiptStore.On("SaveReceipt", mock.Anything).Return(&testUuid, nil)

	//no items
	resUuid, err := receiptService.ProcessReceipt(testReceipt)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "items can not be empty")
	assert.Nil(t, resUuid)

	// invalid total price
	testReceipt.Total = -1
	testReceipt.Items = []model.Item{
		{
			ShortDescription: "description",
			Price:            0,
		},
	}

	resUuid, err = receiptService.ProcessReceipt(testReceipt)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid total price")
	assert.Nil(t, resUuid)

	// invalid purchase date format
	testReceipt.Total = 0
	testReceipt.PurchaseDate = ""
	resUuid, err = receiptService.ProcessReceipt(testReceipt)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid purchase date format")
	assert.Nil(t, resUuid)

	// invalid purchase time format
	testReceipt.Total = 0
	testReceipt.PurchaseDate = "1999-01-02"
	testReceipt.PurchaseTime = "1999-01-02"
	resUuid, err = receiptService.ProcessReceipt(testReceipt)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid purchase time format")
	assert.Nil(t, resUuid)
}

func TestGetReceiptService(t *testing.T) {
	testReceipt := &model.Receipt{
		Retailer:     "TestRetailer",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:22",
		Items:        []model.Item{},
		Total:        0,
	}

	testUuid := uuid.New()

	mockReceiptStore := &MockReceiptStore{}
	receiptService := &ReceiptService{
		receiptStore: mockReceiptStore,
	}
	mockReceiptStore.On("ReadReceipt", mock.Anything).Return(testReceipt, nil)

	receipt, err := receiptService.GetReceipt(testUuid)

	assert.Nil(t, err)

	assert.Equal(t, receipt.Retailer, testReceipt.Retailer)
	assert.Equal(t, len(receipt.Items), len(testReceipt.Items))
	assert.Equal(t, receipt.PurchaseDate, testReceipt.PurchaseDate)
	assert.Equal(t, receipt.PurchaseTime, testReceipt.PurchaseTime)
	assert.Equal(t, receipt.Total, testReceipt.Total)
}

func TestCalculateReceiptPoint(t *testing.T) {

	expectedRetailerPoints := countAlphanNumericCharacters(testReceipt.Retailer)
	// calculate retailer's characters
	points, err := calculatePointsFromReceipt(testReceipt)

	assert.Nil(t, err)
	assert.Equal(t, expectedRetailerPoints, points)

	// 50 points if the total is a round dollar amount with no cents.
	// there's no integer that's not a multiple of 0.25, so it's always
	// gonna add 25 points for being multiple of 0.25
	testReceipt.Total = 12
	expectedPoints := 50 + expectedRetailerPoints

	// 25 points if the total is a multiple of 0.25
	expectedPoints += 25
	points, err = calculatePointsFromReceipt(testReceipt)

	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)

	// 5 points for every two items on the receipt.
	testReceipt.Items = []model.Item{
		{
			ShortDescription: " ",
			Price:            1,
		},
		{
			ShortDescription: "",
			Price:            1,
		},
		{
			ShortDescription: "",
			Price:            1,
		},
		{
			ShortDescription: "",
			Price:            1,
		},
		{
			ShortDescription: "",
			Price:            1,
		},
	}

	expectedPoints += 10
	points, err = calculatePointsFromReceipt(testReceipt)

	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)

	// multiply the price by 0.2 if description length is multiple of 3
	testReceipt.Items[0].ShortDescription = "123"
	testReceipt.Items[1].ShortDescription = " 123 "
	testReceipt.Items[2].ShortDescription = " 456 "

	expectedPoints += int64(math.Ceil(testReceipt.Items[0].Price*0.2) + math.Ceil(testReceipt.Items[1].Price*0.2) + math.Ceil(testReceipt.Items[2].Price*0.2))

	points, err = calculatePointsFromReceipt(testReceipt)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)

	// 6 points if the day in the purchase date is odd.
	testReceipt.PurchaseDate = "1999-01-21"
	expectedPoints += 6
	points, err = calculatePointsFromReceipt(testReceipt)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.

	testReceipt.PurchaseTime = "14:01"
	expectedPoints += 10
	points, err = calculatePointsFromReceipt(testReceipt)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)
}

func TestCalculateReceiptPointTotalNotMultipleOfPoint25(t *testing.T) {
	testReceiptFail := &model.Receipt{
		Retailer:     "",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:00",
		Items:        []model.Item{},
		Total:        12.12,
	}

	// FAIL 25 points if the total is a multiple of 0.25
	testReceiptFail.Total = 1.07
	var expectedPoints int64 = 0
	points, err := calculatePointsFromReceipt(testReceiptFail)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)

}

func TestCalculateReceiptPointTotalNotRoundAmountNorMultipleOfPoint25(t *testing.T) {
	// test case where receipt's total is not a round amount
	// and not multiple of 0.25
	testReceiptFail := &model.Receipt{
		Retailer:     "",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:00",
		Items:        []model.Item{},
		Total:        12.12,
	}
	// FAIL 50 points if the total is a round dollar amount with no cents.
	// FAIL 25 points if the total is a multiple of 0.25
	testReceiptFail.Total = 1.07
	var expectedPoints int64 = 0
	points, err := calculatePointsFromReceipt(testReceiptFail)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)
}

func TestCalculateReceiptPointTotalNotRoundAmountButMultipleOfPoint25(t *testing.T) {
	// test case where receipt's total is not a round amount
	// but is multiple of 0.25
	testReceiptFail := &model.Receipt{
		Retailer:     "",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:00",
		Items:        []model.Item{},
		Total:        1.25,
	}
	// FAIL 50 points if the total is a round dollar amount with no cents.
	// FAIL 25 points if the total is a multiple of 0.25
	var expectedPoints int64 = 25
	points, err := calculatePointsFromReceipt(testReceiptFail)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)
}

func TestCalculateReceiptPointItemNotSatisfy(t *testing.T) {
	testReceiptFail := &model.Receipt{
		Retailer:     "",
		PurchaseDate: "1999-01-20",
		PurchaseTime: "13:00",
		Items:        []model.Item{},
		Total:        0,
	}
	//FAIL 5 points for every two items on the receipt.

	// no items
	testReceiptFail.Items = []model.Item{}
	var expectedPoints int64 = 0
	points, err := calculatePointsFromReceipt(testReceiptFail)
	assert.Nil(t, err)
	assert.Equal(t, expectedPoints, points)
}
