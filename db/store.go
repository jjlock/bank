package db

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	"github.com/jjlock/bank/currency"
	"github.com/jjlock/bank/validate"
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	GetAccount(username string) (account, error)
	AddAccount(username, hashedPassword string, balance currency.MicroUSD) error
	Deposit(username string, amount currency.MicroUSD) error
	Withdraw(username string, amount currency.MicroUSD) error
	Load(file string) error
	Save(file string) error
}

type Memcache map[string]*account

func (m Memcache) GetAccount(username string) (account, error) {
	if err := validate.Username(username); err != nil {
		return account{}, err
	}

	a, ok := m[username]
	if !ok {
		return account{}, errors.New("account not found")
	}

	return *a, nil
}

func (m Memcache) AddAccount(username, password string, balance currency.MicroUSD) error {
	if err := validate.Username(username); err != nil {
		return err
	}

	if err := validate.Password(password); err != nil {
		return err
	}

	if _, ok := m[username]; ok {
		return errors.New("account already exists")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not encrypt password: %w", err)
	}

	m[username] = &account{username, string(bytes), balance}

	return nil
}

func (m Memcache) Deposit(username string, amount currency.MicroUSD) error {
	if err := validate.Username(username); err != nil {
		return err
	}

	a, ok := m[username]
	if !ok {
		return errors.New("account not found")
	}

	a.Balance += amount

	return nil
}

func (m Memcache) Withdraw(username string, amount currency.MicroUSD) error {
	if err := validate.Username(username); err != nil {
		return err
	}

	a, ok := m[username]
	if !ok {
		return errors.New("account not found")
	}

	if amount > a.Balance {
		return errors.New("amount to withdraw is greater than current balance")
	}

	a.Balance -= amount

	return nil
}

func (m Memcache) Load(file string) error {
	db, err := os.Open(file)
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(db)

	err = dec.Decode(&m)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}

func (m Memcache) Save(file string) error {
	db, err := os.Create(file)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(db)

	err = enc.Encode(m)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}
