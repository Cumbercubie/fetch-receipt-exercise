
# RECEIPT PROCESSOR

## TODO:
- Add logging
- Add real DB
- Pre-calculate points for receipt
- Standardize response object
- Tests for handler and DB
## How to run:
- Stand a project root
- Run `go run main.go`


## Description
Go Backend API server to manager receipts and points
### Endpoint: Process Receipts

* Path: `/api/v1/receipts/process`
* Method: `POST`
* Payload: Receipt JSON
* Response: JSON containing an id for the receipt.

Description:

Takes in a JSON receipt and returns a JSON object with an ID generated.

Example Response:
```json
{ "id": "7fb1377b-b223-49d9-a31a-5a02701dd310" }
```

## Endpoint: Get Points

* Path: `/api/v1/receipts/{id}/points`
* Method: `GET`
* Response: A JSON object containing the number of points awarded.

A simple Getter endpoint that looks up the receipt by the ID and returns an object specifying the points awarded.

Example Response:
```json
{ "points": 32 }
```
