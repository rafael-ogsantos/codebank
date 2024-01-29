package usecase

import (
	"codebank/domain"
	"codebank/dto"
	"codebank/infrastructure/kafka"
	"encoding/json"
	"os"
	"time"
)

type UseCaseTransaction struct {
	TransactionRepository domain.TransactionRepository
	KafkaProducer         kafka.KafkaProducer
}

func NewUseCaseTransaction(transactionRepository domain.TransactionRepository) UseCaseTransaction {
	return UseCaseTransaction{TransactionRepository: transactionRepository}
}

func (u UseCaseTransaction) ProcessTransaction(transactionDto dto.Transaction) (*domain.Transaction, error) {
	creditCard := u.hydrateCreditCard(transactionDto)
	ccBalanceLimit, err := u.TransactionRepository.GetCreditCard(creditCard)
	if err != nil {
		return &domain.Transaction{}, err
	}

	creditCard.ID = ccBalanceLimit.ID
	creditCard.Limit = ccBalanceLimit.Limit
	creditCard.Balance = ccBalanceLimit.Balance

	t := u.newTransaction(transactionDto, ccBalanceLimit)
	t.ProcessAndValidate(creditCard)

	err = u.TransactionRepository.SaveTransaction(t, creditCard)
	if err != nil {
		return &domain.Transaction{}, err
	}

	transactionDto.ID = t.ID
	transactionDto.CreatedAt = t.CreatedAt
	transactionJson, err := json.Marshal(transactionDto)
	if err != nil {
		return &domain.Transaction{}, err
	}

	err = u.KafkaProducer.Publish(string(transactionJson), os.Getenv("kafkaTransactionsTopic"))
	if err != nil {
		return &domain.Transaction{}, nil
	}

	return t, nil
}

func (u UseCaseTransaction) hydrateCreditCard(transactionDto dto.Transaction) *domain.CreditCard {
	creditCard := domain.NewCreditCard()
	creditCard.Name = transactionDto.Name
	creditCard.Number = transactionDto.Number
	creditCard.ExpirationMonth = transactionDto.ExpirationMonth
	creditCard.ExpirationYear = transactionDto.ExpirationYear
	creditCard.CVV = transactionDto.CVV
	return creditCard
}

func (u UseCaseTransaction) newTransaction(transactionDto dto.Transaction, cc *domain.CreditCard) *domain.Transaction {
	t := domain.NewTransaction()
	t.CreditCardId = cc.ID
	t.Amount = transactionDto.Amount
	t.Store = transactionDto.Store
	t.Description = transactionDto.Description
	t.CreatedAt = time.Now()
	return t
}
