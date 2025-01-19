package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/glbter/distributed-systems/main-service/cmd"
)

const (
	PORTFOLIO_QUEUE_REQ         = "portfolio_calculation_req"
	PORTFOLIO_QUEUE_RESP        = "portfolio_calculation_resp"
	ORCHESTRATOR_PORTFOLIO_REQ  = "orchestrator_portfolio_req"
	ORCHESTRATOR_PORTFOLIO_RESP = "orchestrator_portfolio_resp"
)

func main() {
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

	if err := initQueues(ch); err != nil {
		log.Fatalln("Failed", err)
	}

	msgs, err := ch.Consume(
		PORTFOLIO_QUEUE_REQ, // queue
		"",                         // consumer
		false,                      // auto-ack
		false,                      // exclusive
		false,                      // no-local
		false,                      // no-wait
		nil,                        // args
	)
	if err != nil {
		log.Fatalln("Failed to initialize a consumer", err)
	}

	logger := cmd.InitLogger()

	logger.Info("server is starting")

	for msg := range msgs {
		go func(msg amqp.Delivery) {
			var (
				start  = time.Now()
				cid    = msg.CorrelationId
				ctx    = context.Background()
				logger = logger.With(zap.String("cid", cid))
			)
			msg.Ack(false)
			logger.Info("start processing of request")

			req := map[string]any{}
			if err := json.Unmarshal(msg.Body, &req); err != nil {
				logger.Error(err.Error())
				respondWithError(ch, err, cid, msg.ReplyTo)
				return
			}

			req["chunk_number"] = 0
			req["iterations_in_chunk"] = 5
			req["elite"] = []any{}

			body, err := json.Marshal(req)
			if err != nil {
				logger.Error(err.Error())
				respondWithError(ch, err, cid, msg.ReplyTo)
				return
			}

			logger.Info("start processing step 1")

			if err := ch.PublishWithContext(ctx,
				"", // exchange
				ORCHESTRATOR_PORTFOLIO_REQ,
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: cid,
					ReplyTo:       ORCHESTRATOR_PORTFOLIO_RESP,
					Body:          body,
					Priority:      msg.Priority,
				}); err != nil {
				logger.Error(err.Error())
				respondWithError(ch, err, cid, msg.ReplyTo)
				return
			}

			respStep1, err := ch.Consume(
				ORCHESTRATOR_PORTFOLIO_RESP, // queue
				"",                         // consumer
				false,                      // auto-ack
				false,                      // exclusive
				false,                      // no-local
				false,                      // no-wait
				nil,                        // args
			)
			if err != nil {
				log.Fatalln("Failed to initialize a consumer", err)
				return
			}

			getResponse := func(responseCh <-chan amqp.Delivery) amqp.Delivery {
				for rs := range responseCh {
					logger.Info(fmt.Sprintf("got message with cid %v", rs.CorrelationId))
					if rs.CorrelationId == cid {
						if err := rs.Ack(false); err != nil {
							logger.Error(err.Error())
						}
						return rs
					} else {
						rs.Reject(true)
					}
				}

				return amqp.Delivery{}
			}

			logger.Info("await finish step 1")

			initStepResp := getResponse(respStep1)

			logger.Info("finished step 1")

			initStepReq := map[string]any{}
			if err := json.Unmarshal(initStepResp.Body, &initStepReq); err != nil {
				logger.Error(err.Error())
				respondWithError(ch, err, cid, msg.ReplyTo)
				return
			}

			req["chunk_number"] = 1
			req["iterations_in_chunk"] = 5
			req["elite"] = initStepReq["elite"]

			body, err = json.Marshal(req)
			if err != nil {
				logger.Error(err.Error())
				respondWithError(ch, err, cid, msg.ReplyTo)
				return
			}

			logger.Info("start processing step 2")

			if err := ch.PublishWithContext(ctx,
				"", // exchange
				ORCHESTRATOR_PORTFOLIO_REQ,
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: cid,
					ReplyTo:       ORCHESTRATOR_PORTFOLIO_RESP,
					Body:          body,
					Priority:      msg.Priority,
				}); err != nil {
				logger.Error(err.Error())
				respondWithError(ch, err, cid, msg.ReplyTo)
				return
			}

			logger.Info("await finish step 2")

			finalStepResp := getResponse(respStep1)

			logger.Info("finished step 2")
			
			finalResp := map[string]any{}
			if err := json.Unmarshal(finalStepResp.Body, &finalResp); err != nil {
				logger.Error(err.Error())
				return
			}

			delete(finalResp, "elite")
			body, err = json.Marshal(finalResp)
			if err != nil {
				logger.Error(err.Error())
				return
			}

			if err := ch.PublishWithContext(ctx,
				"", // exchange
				msg.ReplyTo,
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: cid,
					Body:          body,
					Priority:      msg.Priority,
				}); err != nil {
				logger.Error(err.Error())
				return
			}

			logger.Info("finish", zap.Duration("duration", time.Since(start)))
		}(msg)
	}
}

func respondWithError(ch *amqp.Channel, err error, cid, replyTo string) error {
	resp := map[string]any{}
	resp["error"] = err.Error()
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return ch.Publish(
		"", // exchange
		PORTFOLIO_QUEUE_RESP,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: cid,
			ReplyTo:       replyTo,
			Body:          body,
			Priority:      1,
		},
	)
}

func initQueues(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(
		PORTFOLIO_QUEUE_REQ, // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // noWait
		nil,                 // arguments
	); err != nil {
		return fmt.Errorf("declare a queue for calculation request: %w", err)
	}

	if _, err := ch.QueueDeclare(
		PORTFOLIO_QUEUE_RESP, // name
		false,                // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // noWait
		nil,                  // arguments
	); err != nil {
		return fmt.Errorf("declare a queue for calculation response: %w", err)
	}

	if _, err := ch.QueueDeclare(
		ORCHESTRATOR_PORTFOLIO_REQ, // name
		false,                      // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // noWait
		nil,                        // arguments
	); err != nil {
		return fmt.Errorf("declare a queue for request from orchestrator: %w", err)
	}

	if _, err := ch.QueueDeclare(
		ORCHESTRATOR_PORTFOLIO_RESP, // name
		false,                       // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // noWait
		nil,                         // arguments
	); err != nil {
		return fmt.Errorf("declare a queue for response to orchestrator: %w", err)
	}

	return nil
}
