import time as time
import logging
import sys
import json

from flask import Flask, request
from flask_restful import Resource, Api

import pandas as pd

from engine import Engine


app = Flask(__name__)
logger = logging.getLogger(__name__)
api = Api(app)
port = 8080

app.logger.setLevel(logging.INFO)

logging.basicConfig(
    format='%(asctime)s %(levelname)-8s %(message)s',
    level=logging.INFO,
    datefmt='%Y-%m-%d %H:%M:%S')


class StockDateData(object):
    def __init__(self, date, close):
        self.date = date
        self.close = close

    def to_dict(self):
        return {
            'date': self.date,
            'close': self.close,
        }

class RecommendationEngine(Resource):
    def __init__(self):
        self.engine = Engine(logger)

    def stockToJson(self, jData):
        stocks = []
        for item in jData['data']:
            stock = {"name":None, "data":None}
            stock['name'] = item['name']

            stockData = []
            for item in item['data']:
                stockData.append(StockDateData(**item))

            stock['data'] = stockData
            stocks.append(stock)


    def post(self):
        start = time.time()

        j = request.get_json()

        dfs = []
        for stock in j['data']:
            stockData = []
            for item in stock['data']:
                s = StockDateData(**item)
                stockData.append((s.date, s.close))
            
            dfs.append(pd.DataFrame(stockData, columns=['Date', stock['name']]))
           
        returns, risk, portfolio = self.engine.create_portfolio(dfs, len(dfs))

        logger.info(f"execution time {time.time() - start}")
       
        return {
            'returns': returns,
            'risk': risk,
            'portfolio': portfolio.to_dict(),
        }, 200


api.add_resource(RecommendationEngine, '/recommendation/run')

if __name__ == '__main__':
    if sys.argv.__len__() > 1:
        port = sys.argv[1]

    logger.info(f"api starting on port {port}")

    app.run(host="0.0.0.0", port=port, debug=True)
