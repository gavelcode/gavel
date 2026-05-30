import logging

from ecommerce.application.order.order_dto import OrderDto
from ecommerce.domain.order.order import Order
from ecommerce.domain.order.order_line import OrderLine


logger = logging.getLogger(__name__)


class PlaceOrderHandler:
    def __init__(self, order_repo, customer_repo, product_repo, inventory_repo):
        self._order_repo = order_repo
        self._customer_repo = customer_repo
        self._product_repo = product_repo
        self._inventory_repo = inventory_repo

    def execute(self, customer_id, items):
        customer = self._customer_repo.find_by_id(customer_id)
        if customer is None:
            raise ValueError(f"customer not found: {customer_id}")

        if not items:
            raise ValueError("order must have at least one item")

        order_id = self._order_repo.next_id()
        order = Order(order_id, customer.id)

        for item in items:
            if item["quantity"] <= 0:
                raise ValueError(f"quantity must be positive for product {item['product_id']}")

            product = self._product_repo.find_by_id(item["product_id"])
            if product is None:
                raise ValueError(f"product not found: {item['product_id']}")

            stock = self._inventory_repo.find_by_product_id(item["product_id"])
            if stock is None:
                raise ValueError(f"no stock for product: {item['product_id']}")

            if not stock.has_enough(item["quantity"]):
                raise RuntimeError(
                    f"insufficient stock for {product.name}: "
                    f"available={stock.quantity}, requested={item['quantity']}"
                )

            stock.reserve(item["quantity"])
            self._inventory_repo.save(stock)

            line = OrderLine(product.id, product.name, item["quantity"], product.price)
            order.add_line(line)

        total = order.total()
        if total.is_zero():
            raise RuntimeError("order total cannot be zero")

        order.confirm()
        self._order_repo.save(order)

        # DELIBERATE: f-string in logging (Ruff G004)
        logging.info(f"Order {order.id} placed for customer {customer_id} with total {total}")

        return OrderDto.from_domain(order)
