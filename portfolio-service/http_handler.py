import pandas as pd
import time

from entities import StockDateData
from engine import Engine


from flask_restful import Resource, Api
from flask import Flask, request



class HttpRecommendationEngine(Resource):
    def __init__(self, logger):
        self.engine = Engine(logger)
        self.logger = logger

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

        self.logger.info(f"execution time {time.time() - start}")

        return {
            'returns': returns,
            'risk': risk,
            'portfolio': portfolio.to_dict(),
        }, 200

