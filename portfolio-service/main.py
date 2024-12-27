import time as time
import logging
import sys

from flask import Flask
from flask_restful import Resource, Api

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

class RecommendationEngine(Resource):
    def __init__(self):
        self.engine = Engine(logger)

    def post(self):
        start = time.time()

        returns, risk, portfolio = self.engine.create_portfolio(6)

        logger.info(f"execution time {time.time() - start}")
        
        return {
            'returns': returns,
            'risk': risk,
            'portfolio': portfolio.to_dict(),
        }, 200


api.add_resource(RecommendationEngine, '/run')

if __name__ == '__main__':
    if sys.argv.__len__() > 1:
        port = sys.argv[1]

    logger.info(f"api starting on port {port}")

    app.run(host="0.0.0.0", port=port, debug=True)
