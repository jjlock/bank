package db

import "github.com/jjlock/bank/server/currency"

type account struct {
	Username     string
	PasswordHash string
	Balance      currency.MicroUSD
}
