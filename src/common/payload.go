package common

// order

type CreateOrderResponse struct {
	OrderId string `json:"order_id"`
}

type FindOrderResponse struct {
	OrderId   string   `json:"order_id"`
	Paid      bool     `json:"paid"`
	Items     []string `json:"items"`
	UserId    string   `json:"user_id"`
	TotalCost int      `json:"total_cost"`
}

// stock

type CreateItemResponse struct {
	ItemId string `json:"item_id"`
}

type FindItemResponse struct {
	Price int `json:"price"`
	Stock int `json:"stock"`
}

// payment

type CreateUserResponse struct {
	UserId string `json:"user_id"`
}

type FindUserResponse struct {
	UserId string `json:"user_id"`
	Credit int    `json:"credit"`
}

type AddFundsResponse struct {
	Done bool `json:"done"`
}

type PaymentStatusResponse struct {
	Paid bool `json:"paid"`
}

// transaction checkout

type IdAmountPair struct {
	Id     string `json:"id"` // item_id or user_id
	Amount int    `json:"amount"`
}

type ItemTxPrpRequest struct {
	TxId  string         `json:"tx_id"`
	Items []IdAmountPair `json:"items"`
}

type ItemTxPrpResponse struct {
	TotalCost int `json:"total_cost"`
}

type CreditTxPrpRequest struct {
	TxId  string       `json:"tx_id"`
	Payer IdAmountPair `json:"payer"`
}
