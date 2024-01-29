package main

import (
	"codebank/infrastructure/grpc/server"
	"codebank/infrastructure/kafka"
	"codebank/infrastructure/repository"
	"codebank/usecase"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
}

func main() {
	db := setupDb()
	defer db.Close()
	producer := setupKafkaProducer()
	processTransactionUseCase := setupTransactionUseCase(db, producer)
	serveGrpc(processTransactionUseCase)
}

func setupTransactionUseCase(db *sql.DB, producer kafka.KafkaProducer) usecase.UseCaseTransaction {
	transactionRepository := repository.NewTransactionRepositoryDb(db)
	useCase := usecase.NewUseCaseTransaction(transactionRepository)
	useCase.KafkaProducer = producer
	return useCase
}

func setupKafkaProducer() kafka.KafkaProducer {
	producer := kafka.NewKafkaProducer()
	err := producer.SetupProducer(os.Getenv("kafkaBootstrapServers"))
	if err != nil {
		log.Fatalf("Failed to set up producer: %v", err)
	}

	return producer
}

func setupDb() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("host"),
		os.Getenv("port"),
		os.Getenv("user"),
		os.Getenv("password"),
		os.Getenv("dbname"),
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("error connection to database")
	}

	return db
}

func serveGrpc(processTransactionUseCase usecase.UseCaseTransaction) {
	grpcServer := server.NewGRPCServer()
	grpcServer.ProcessTransactionUseCase = processTransactionUseCase
	fmt.Println("rodando gRPC Server")
	grpcServer.Serve()
}

// cc := domain.NewCreditCard()
// 	cc.Name = "MAU"
// 	cc.Number = "2121222"
// 	cc.ExpirationMonth = 12
// 	cc.ExpirationYear = 2023
// 	cc.CVV = 188
// 	cc.Limit = 1200
// 	cc.Balance = 0

// 	repo := repository.NewTransactionRepositoryDb(db)
// 	err := repo.CreateCreditCard(cc)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
