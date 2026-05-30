class Stock:
    def __init__(self, product_id, quantity):
        if product_id <= 0:
            raise ValueError("product id must be positive")
        if quantity < 0:
            raise ValueError("stock quantity cannot be negative")

        self._product_id = product_id
        self._quantity = quantity

    @property
    def product_id(self):
        return self._product_id

    @property
    def quantity(self):
        return self._quantity

    def has_enough(self, requested):
        return self._quantity >= requested

    def reserve(self, amount):
        if amount <= 0:
            raise ValueError("reserve amount must be positive")
        if not self.has_enough(amount):
            raise RuntimeError(
                f"insufficient stock for product {self._product_id}: "
                f"available={self._quantity}, requested={amount}"
            )
        self._quantity -= amount

    def restock(self, amount):
        if amount <= 0:
            raise ValueError("restock amount must be positive")
        self._quantity += amount
