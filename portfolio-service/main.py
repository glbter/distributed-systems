import time as time
import logging
import sys
import json

from flask_restful import Resource, Api
from flask import Flask, request


# app = Flask(__name__)
logger = logging.getLogger(__name__)
# api = Api(app)
# port = 8080

# app.logger.setLevel(logging.INFO)

logging.basicConfig(
    format='%(asctime)s %(levelname)-8s %(message)s',
    level=logging.INFO,
    datefmt='%Y-%m-%d %H:%M:%S')




# api.add_resource(RecommendationEngine, '/recommendation/run')

if __name__ == '__main__':
    if sys.argv.__len__() > 1:
        port = sys.argv[1]

    logger.info(f"api starting on port {port}")

    # app.run(host="0.0.0.0", port=port, debug=True)
