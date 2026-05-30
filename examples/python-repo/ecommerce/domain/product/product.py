class Product:
    def __init__(self, product_id, name, description, price):
        if product_id <= 0:
            raise ValueError("product id must be positive")
        if not name or not name.strip():
            raise ValueError("product name must not be blank")
        if price is None:
            raise ValueError("price must not be None")

        self._id = product_id
        self._name = name
        self._description = description or ""
        self._price = price

    @property
    def id(self):
        return self._id

    @property
    def name(self):
        return self._name

    @property
    def description(self):
        return self._description

    @property
    def price(self):
        return self._price
