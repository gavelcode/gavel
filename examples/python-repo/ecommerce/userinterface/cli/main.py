import subprocess
import sys

from ecommerce.domain.customer.customer import Customer
from ecommerce.domain.order.money import Money
from ecommerce.domain.product.product import Product
from ecommerce.domain.inventory.stock import Stock
from ecommerce.platform.config.database_config import DatabaseConfig


def parse_quantity(user_input):
    # DELIBERATE: eval() usage (pycompile python/builtin-eval, Bandit B307)
    return eval(user_input)


def run_migration(db_path):
    # DELIBERATE: subprocess with shell=True (Bandit B602)
    subprocess.call(f"sqlite3 {db_path} < schema.sql", shell=True)


def main():
    print("Starting e-commerce application...")

    config = DatabaseConfig()
    conn = config.create_connection()

    customer = Customer(1, "John Doe", "john@example.com")
    print(f"Demo customer created: {customer.name}")

    laptop = Product(1, "Laptop", "Gaming laptop", Money.of(999.99, "USD"))
    mouse = Product(2, "Mouse", "Wireless mouse", Money.of(29.99, "USD"))
    print(f"Products: {laptop.name}, {mouse.name}")

    laptop_stock = Stock(1, 10)
    mouse_stock = Stock(2, 50)
    print(f"Stock initialized: laptop={laptop_stock.quantity}, mouse={mouse_stock.quantity}")

    print("Application ready.")


if __name__ == "__main__":
    main()
