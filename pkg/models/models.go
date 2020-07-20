package models

import (
	"strconv"
	"time"
)

type Wallet struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type Send struct {
	SenderId    string `json:"sender-id"`
	RecipientId string `json:"recipient-id"`
	Amount      uint64 `json:"amount"`
}

type Topup struct {
	RecipientId string `json:"recipient-id"`
	Amount      uint64 `json:"amount"`
}

type Transaction struct {
	TransactionId int64
	SenderId      string
	RecipientId   string
	Amount        uint64
	Date          time.Time
}

func (t *Transaction) ToSliceOfStrings() []string {
	var strs []string
	strs = append(strs, strconv.FormatInt(t.TransactionId, 10))
	strs = append(strs, t.SenderId)
	strs = append(strs, t.RecipientId)
	strs = append(strs, strconv.FormatUint(t.Amount, 10))
	strs = append(strs, t.Date.Format("2006-01-02 15:04:05 -0700"))
	return strs
}
