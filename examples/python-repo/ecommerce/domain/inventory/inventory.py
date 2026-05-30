class Inventory:
    def __init__(self):
        self._stocks = {}

    def add_stock(self, stock):
        self._stocks[stock.product_id] = stock

    def get_stock(self, product_id):
        stock = self._stocks.get(product_id)
        if stock is None:
            raise ValueError(f"no stock record for product {product_id}")
        return stock

    def is_available(self, product_id, quantity):
        stock = self._stocks.get(product_id)
        return stock is not None and stock.has_enough(quantity)

    def reserve(self, product_id, quantity):
        self.get_stock(product_id).reserve(quantity)

    def restock(self, product_id, quantity):
        self.get_stock(product_id).restock(quantity)
