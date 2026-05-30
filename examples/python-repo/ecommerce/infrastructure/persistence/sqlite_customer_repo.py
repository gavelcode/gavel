from ecommerce.domain.customer.customer import Customer
from ecommerce.domain.customer.customer_repository import CustomerRepository

# DELIBERATE: hardcoded password (Bandit B105)
DB_PASSWORD = "s3cret_passw0rd"


class SqliteCustomerRepository(CustomerRepository):
    def __init__(self, connection):
        self._conn = connection

    def save(self, customer):
        self._conn.execute(
            "INSERT OR REPLACE INTO customers (id, name, email) VALUES (?, ?, ?)",
            (customer.id, customer.name, customer.email),
        )
        self._conn.commit()

    def find_by_id(self, customer_id):
        cursor = self._conn.execute(
            "SELECT id, name, email FROM customers WHERE id = ?",
            (customer_id,),
        )
        row = cursor.fetchone()
        if row is None:
            return None
        return Customer(row[0], row[1], row[2])
