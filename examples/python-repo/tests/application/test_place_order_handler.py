import unittest

from ecommerce.application.order.place_order_handler import PlaceOrderHandler
from ecommerce.domain.customer.customer import Customer
from ecommerce.domain.inventory.stock import Stock
from ecommerce.domain.order.money import Money
from ecommerce.domain.order.order import Order
from ecommerce.domain.product.product import Product


class InMemoryOrderRepo:
    def __init__(self):
        self._orders = {}
        self._counter = 0

    def save(self, order):
        self._orders[order.id] = order

    def find_by_id(self, order_id):
        return self._orders.get(order_id)

    def next_id(self):
        self._counter += 1
        return self._counter


class InMemoryCustomerRepo:
    def __init__(self):
        self._customers = {}

    def save(self, customer):
        self._customers[customer.id] = customer

    def find_by_id(self, customer_id):
        return self._customers.get(customer_id)


class InMemoryProductRepo:
    def __init__(self):
        self._products = {}

    def add(self, product):
        self._products[product.id] = product

    def find_by_id(self, product_id):
        return self._products.get(product_id)


class InMemoryInventoryRepo:
    def __init__(self):
        self._stocks = {}

    def save(self, stock):
        self._stocks[stock.product_id] = stock

    def find_by_product_id(self, product_id):
        return self._stocks.get(product_id)


class TestPlaceOrderHandler(unittest.TestCase):
    def setUp(self):
        self.order_repo = InMemoryOrderRepo()
        self.customer_repo = InMemoryCustomerRepo()
        self.product_repo = InMemoryProductRepo()
        self.inventory_repo = InMemoryInventoryRepo()

        self.customer_repo.save(Customer(1, "Alice", "alice@example.com"))
        self.product_repo.add(Product(1, "Laptop", "Gaming laptop", Money.of(999.99, "USD")))
        self.product_repo.add(Product(2, "Mouse", "Wireless mouse", Money.of(29.99, "USD")))
        self.inventory_repo.save(Stock(1, 10))
        self.inventory_repo.save(Stock(2, 50))

        self.handler = PlaceOrderHandler(
            self.order_repo, self.customer_repo, self.product_repo, self.inventory_repo
        )

    def test_place_order_successfully(self):
        items = [
            {"product_id": 1, "quantity": 2},
            {"product_id": 2, "quantity": 3},
        ]
        result = self.handler.execute(1, items)
        self.assertIsNotNone(result)
        self.assertEqual(result.customer_id, 1)
        self.assertEqual(result.status, "confirmed")
        self.assertEqual(len(result.lines), 2)

    def test_reject_unknown_customer(self):
        with self.assertRaises(ValueError):
            self.handler.execute(999, [{"product_id": 1, "quantity": 1}])

    def test_reject_empty_items(self):
        with self.assertRaises(ValueError):
            self.handler.execute(1, [])

    def test_reject_insufficient_stock(self):
        with self.assertRaises(RuntimeError):
            self.handler.execute(1, [{"product_id": 1, "quantity": 100}])


if __name__ == "__main__":
    unittest.main()
