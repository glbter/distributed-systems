package rabbit

import (
	"context"
	"encoding/json"
	"github.com/glbter/distributed-systems/main-service/entities"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	PORTFOLIO_QUEUE_REQ = "portfolio_calculation_req"
	PORTFOLIO_QUEUE_RESP = "portfolio_calculation_resp"
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
	body, err := json.Marshal(map[string]any{
		"data": data,
	})
	if err != nil {
		return err
	}

	return c.channel.PublishWithContext(ctx,
		"",          // exchange
		PORTFOLIO_QUEUE_REQ,
		// "rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: cid,
			ReplyTo:       PORTFOLIO_QUEUE_RESP,
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
		PORTFOLIO_QUEUE_RESP, // queue
		"",              // consumer
		false,            // auto-ack
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
