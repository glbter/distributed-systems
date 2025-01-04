import pika

PORTFOLIO_QUEUE = "portfolio_calculation"

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host="localhost"),
)
channel = connection.channel()

channel.queue_declare(queue="PORTFOLIO_QUEUE")

channel.basic_publish(exchange="", routing_key="rpc_queue", body="Hello World!")
print(" [x] Sent 'Hello World!'")

connection.close()