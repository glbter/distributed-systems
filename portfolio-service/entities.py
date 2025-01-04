class StockDateData(object):
    def __init__(self, date, close):
        self.date = date
        self.close = close

    def to_dict(self):
        return {
            'date': self.date,
            'close': self.close,
        }
