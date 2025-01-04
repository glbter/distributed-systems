package rabbit

import (
	"context"
	"encoding/json"
	"github.com/glbter/distributed-systems/main-service/entities"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	PORTFOLIO_QUEUE = "portfolio_calculation"
)

func NewPortfolioServiceClient(channel *amqp.Channel) *PortfolioOptimizatiorClient {
	return &PortfolioOptimizatiorClient{
		channel: channel,
	}
}

type PortfolioOptimizatiorClient struct {
	channel *amqp.Channel
}

func (c *PortfolioOptimizatiorClient) StartRecommend(ctx context.Context, data []entities.StocksHistory, cid string, isVipUser bool) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.channel.PublishWithContext(ctx,
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: cid,
			ReplyTo:       PORTFOLIO_QUEUE,
			Body:          body,
			Priority:      c.userPriority(isVipUser), // higher prio for convolution
		})

}

func (c *PortfolioOptimizatiorClient) userPriority(isVip bool) uint8 {
	if isVip {
		return 3
	}

	return 1
}

func (c *PortfolioOptimizatiorClient) ReceiveRecommend() (<-chan amqp.Delivery, error) {
	msgs, err := c.channel.Consume(
		PORTFOLIO_QUEUE, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
