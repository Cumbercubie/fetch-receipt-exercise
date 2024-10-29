package db

import (
	"errors"

	"github.com/google/uuid"
	"github.com/receipt-processor/receipt/model"
)

type dbHandler interface {
	SaveReceipt(receipt model.Receipt) (*uuid.UUID, error)
}
type DBHandler struct {
	receiptStoreMap map[uuid.UUID]*model.Receipt
}

func InitDB() *DBHandler {
	return &DBHandler{
		receiptStoreMap: make(map[uuid.UUID]*model.Receipt),
	}
}

func (db *DBHandler) SaveReceipt(receipt *model.Receipt) (*uuid.UUID, error) {
	newUUuid := uuid.New()
	db.receiptStoreMap[newUUuid] = receipt
	return &newUUuid, nil
}

func (db *DBHandler) ReadReceipt(id uuid.UUID) (*model.Receipt, error) {
	foundReceipt, ok := db.receiptStoreMap[id]

	if !ok {
		return nil, errors.New("cannot find receipt with this ID")
	}

	return foundReceipt, nil
}
