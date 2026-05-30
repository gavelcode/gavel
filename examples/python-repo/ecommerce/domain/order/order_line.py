from ecommerce.domain.order.money import Money


class OrderLine:
    def __init__(self, product_id, product_name, quantity, unit_price):
        if not product_name:
            raise ValueError("product name must not be empty")
        if quantity <= 0:
            raise ValueError("quantity must be positive")
        if unit_price is None:
            raise ValueError("unit price must not be None")

        self._product_id = product_id
        self._product_name = product_name
        self._quantity = quantity
        self._unit_price = unit_price

    @property
    def product_id(self):
        return self._product_id

    @property
    def product_name(self):
        return self._product_name

    @property
    def quantity(self):
        return self._quantity

    @property
    def unit_price(self):
        return self._unit_price

    def line_total(self):
        return self._unit_price.multiply(self._quantity)
