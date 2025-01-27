#!/usr/bin/env python
import os
import sys
import pika
import time
import json
import os
import logging

import pandas as pd

from engine import Engine
from entities import StockDateData

PORTFOLIO_QUEUE_REQ = "portfolio_calculation_req"
PORTFOLIO_QUEUE_RESP = "portfolio_calculation_resp"

class AsyncRecommendationEngine():
    def __init__(self, logger, channel):
        self.engine = Engine(logger)
        self.channel = channel
        self.logger = logger


    def portfolioCallback(self, ch, method, properties, body):
        start = time.time()
        j = json.loads(body.decode())

        dfs = []
        for stock in j['data']:
            stockData = []
            for item in stock['data']:
                s = StockDateData(**item)
                stockData.append((s.date, s.close))

            dfs.append(pd.DataFrame(stockData, columns=['Date', stock['name']]))

        returns, risk, portfolio = self.engine.create_portfolio(dfs, len(dfs))
        self.channel.basic_publish(
            exchange="",
            routing_key=properties.reply_to,
            properties=pika.BasicProperties(correlation_id=properties.correlation_id),
            body=json.dumps({
                'returns': returns,
                'risk': risk,
                'portfolio': portfolio.to_dict(),
            }),
        )

        ch.basic_ack(delivery_tag=method.delivery_tag)

        self.logger.info(f"cid {properties.correlation_id} execution time {time.time() - start}")


def mainAsync(logger):
    rabbitUrl = os.environ['RABBIT_URL_PORTFOLIO']
    if rabbitUrl == "":
        print("rabbit URL not set")
        sys.exit(1)

    connection = pika.BlockingConnection(
        pika.URLParameters(rabbitUrl),
    )
    channel = connection.channel()

    channel.queue_declare(queue=PORTFOLIO_QUEUE_REQ)
    channel.queue_declare(queue=PORTFOLIO_QUEUE_RESP)

    handler = AsyncRecommendationEngine(logger, channel)

    channel.basic_consume(
        queue=PORTFOLIO_QUEUE_REQ,
        on_message_callback=handler.portfolioCallback,
    )

    channel.basic_qos(prefetch_count=1)
    print(" [*] Waiting for messages. To exit press CTRL+C")
    channel.start_consuming()


if __name__ == "__main__":
    try:
        logger = logging.getLogger(__name__)

        logging.basicConfig(
            format='%(asctime)s %(levelname)-8s %(message)s',
            level=logging.INFO,
            datefmt='%Y-%m-%d %H:%M:%S')

        mainAsync(logger)
    except KeyboardInterrupt:
        print("Interrupted")
        try:
            sys.exit(0)
        except SystemExit:
            os._exit(0)