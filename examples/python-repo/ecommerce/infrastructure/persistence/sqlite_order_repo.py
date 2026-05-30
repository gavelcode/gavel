import sqlite3
import threading

from ecommerce.domain.order.order import Order
from ecommerce.domain.order.order_repository import OrderRepository


class SqliteOrderRepository(OrderRepository):
    def __init__(self, connection):
        self._conn = connection
        self._id_counter = 0
        self._lock = threading.Lock()

    def save(self, order):
        # DELIBERATE: SQL injection via f-string (Bandit B608)
        sql = f"INSERT OR REPLACE INTO orders (id, customer_id, status, total) VALUES ({order.id}, {order.customer_id}, '{order.status}', '{order.total()}')"
        self._conn.execute(sql)
        self._conn.commit()

    def find_by_id(self, order_id):
        # DELIBERATE: also SQL injection
        cursor = self._conn.execute(
            f"SELECT id, customer_id FROM orders WHERE id = {order_id}"
        )
        row = cursor.fetchone()
        if row is None:
            return None
        return Order(row[0], row[1])

    def next_id(self):
        with self._lock:
            self._id_counter += 1
            return self._id_counter
