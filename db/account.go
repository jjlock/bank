package db

import "github.com/jjlock/bank/currency"

type account struct {
	Username     string
	PasswordHash string
	Balance      currency.MicroUSD
}
