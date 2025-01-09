package cmd

import (
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"

	portfolioHttp "github.com/glbter/distributed-systems/main-service/http"
	"github.com/glbter/distributed-systems/main-service/optimizer/client/rabbit"
	"github.com/glbter/distributed-systems/main-service/optimizer/repo/csv"
)

func ExecuteAsync(logger *zap.Logger) {
	//logger := InitLogger()

	rabbitUrl := os.Getenv("RABBIT_URL_PORTFOLIO_ENGINE")
	if rabbitUrl == "" {
		log.Fatalln("rabbit url is empty")
	}

	conn, err := amqp.Dial(rabbitUrl)
	if err != nil {
		log.Fatalln("Failed to connect to RabbitMQ", err)
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalln("Failed to open a channel", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		rabbit.PORTFOLIO_QUEUE_REQ, // name
		false,                  // durable
		false,                  // delete when unused
		false,                   // exclusive
		false,                  // noWait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatalln("Failed to declare a queue", err)
	}

	_, err = ch.QueueDeclare(
		rabbit.PORTFOLIO_QUEUE_RESP, // name
		false,                  // durable
		false,                  // delete when unused
		false,                   // exclusive
		false,                  // noWait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatalln("Failed to declare a queue", err)
	}

	client := rabbit.NewPortfolioServiceClient(ch)

	msgCh, err := client.ReceiveRecommend()
	if err != nil {
		log.Fatalln("Failed to initialize a portfolio server", err)
	}

	//httpClient := optimizerHttp.NewClient(client, url, logger)
	handler := portfolioHttp.PortfolioHandler{
		//PortfolioEngine: &httpClient,
		StockRepo:            csv.StockRepo{},
		Logger:               logger,
		PortfolioEngineAsync: client,
		ReplyMessageCh:       msgCh,
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/portfolio/optimize", handler.OptimizePortfolioAsync)

	logger.Info("server is starting")
	http.ListenAndServe(":8080", r)
}
