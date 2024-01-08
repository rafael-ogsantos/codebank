package repository

import (
	"codebank/domain"
	"database/sql"
	"errors"
	"fmt"
)

type TransactionRepositoryDb struct {
	db *sql.DB
}

func NewTransactionRepositoryDb(db *sql.DB) *TransactionRepositoryDb {
	return &TransactionRepositoryDb{db: db}
}

func (t *TransactionRepositoryDb) GetCreditCard(creditCard *domain.CreditCard) (*domain.CreditCard, error) {
	var c domain.CreditCard
	stmt, err := t.db.Prepare("select id, balance, balance_limit from credit_cards where number=$1")
	if err != nil {
		return &c, err
	}
	if err = stmt.QueryRow(creditCard.Number).Scan(&c.ID, &c.Balance, &c.Limit); err != nil {
		return &c, errors.New("credit card does not exists")
	}
	return &c, nil
}

func (t *TransactionRepositoryDb) SaveTransaction(transaction *domain.Transaction, creditCard *domain.CreditCard) error {
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		insert into transactions(id, credit_card_id, amount, status, description, store, created_at)
		values($1, $2, $3, $4, $5, $6, $7)
	`,
		transaction.ID,
		transaction.CreditCardId,
		transaction.Amount,
		transaction.Status,
		transaction.Description,
		transaction.Store,
		transaction.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	if transaction.Status == "approved" {
		err = t.updateBalance(creditCard)
		if err != nil {
			return nil
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (t *TransactionRepositoryDb) CreateCreditCard(creditCard *domain.CreditCard) error {
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        insert into credit_cards(id, name, number, expiration_month, expiration_year, CVV, balance, balance_limit)
        values($1, $2, $3, $4, $5, $6, $7, $8)
    `,
		creditCard.ID,
		creditCard.Name,
		creditCard.Number,
		creditCard.ExpirationMonth,
		creditCard.ExpirationYear,
		creditCard.CVV,
		creditCard.Balance,
		creditCard.Limit,
	)

	fmt.Println(&creditCard.Balance)

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *TransactionRepositoryDb) updateBalance(creditCard *domain.CreditCard) error {
	_, err := t.db.Exec("update credit_cards set balance = $1 where id = $2",
		creditCard.Balance, creditCard.ID)

	if err != nil {
		return err
	}

	return nil
}
