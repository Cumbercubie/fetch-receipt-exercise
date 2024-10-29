package model

import (
	"errors"
	"fmt"
	"regexp"
)


type Receipt struct {
	Retailer     string  `json:"retailer"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Items        []Item  `json:"items"`
	Total        float64 `json:"total,string"`
}

type Item struct {
	ShortDescription string  `json:"shortDescription"`
	Price            float64 `json:"price,string"`
}

func (r *Receipt) Validate() error {
	retailerPattern := "^[\\w\\s\\-&]+$"
	match, err := regexp.MatchString(retailerPattern, r.Retailer)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("invalid retailer format")
	}

	//format "6.49"
	totalPattern := "^\\d+\\.\\d{2}$"
	totalString := fmt.Sprintf("%.2f", r.Total) // Convert total to string with 2 decimal points
	match, err = regexp.MatchString(totalPattern, totalString)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("invalid total format")
	}

	// Validate items array (must contain at least one item)
	if len(r.Items) < 1 {
		return errors.New("must contain at least one item")
	}

	return nil
}
