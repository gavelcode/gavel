package com.example.ecommerce.userinterface.cli;

import com.example.ecommerce.application.order.PlaceOrderHandler;
import com.example.ecommerce.application.order.OrderDto;
import com.example.ecommerce.application.payment.ProcessPaymentHandler;
import com.example.ecommerce.application.payment.PaymentDto;
import com.example.ecommerce.domain.customer.Customer;
import com.example.ecommerce.domain.inventory.Stock;
import com.example.ecommerce.domain.order.Money;
import com.example.ecommerce.domain.product.Product;
import com.example.ecommerce.infrastructure.persistence.JdbcCustomerRepository;
import com.example.ecommerce.infrastructure.persistence.JdbcOrderRepository;
import com.example.ecommerce.platform.config.DatabaseConfig;
import java.sql.Connection;
import java.util.List;

public class Main {

    // DELIBERATE: hardcoded password (PMD HardCodedCryptoKey)
    private static final String DB_PASSWORD = "s3cret_passw0rd";

    public static void main(String[] args) {
        // DELIBERATE: System.out.println instead of logger (PMD SystemPrintln)
        System.out.println("Starting e-commerce application...");
        System.out.println("Database password configured: " + (DB_PASSWORD != null));

        DatabaseConfig config = new DatabaseConfig();
        Connection conn = config.createConnection();

        JdbcOrderRepository orderRepo = new JdbcOrderRepository(conn);
        JdbcCustomerRepository customerRepo = new JdbcCustomerRepository(conn);

        Customer demoCustomer = new Customer(1, "John Doe", "john@example.com");
        customerRepo.save(demoCustomer);

        System.out.println("Demo customer created: " + demoCustomer.getName());

        Product laptop = new Product(1, "Laptop", "Gaming laptop", Money.of(999.99, "USD"));
        Product mouse = new Product(2, "Mouse", "Wireless mouse", Money.of(29.99, "USD"));

        System.out.println("Products: " + laptop.getName() + ", " + mouse.getName());

        Stock laptopStock = new Stock(1, 10);
        Stock mouseStock = new Stock(2, 50);

        System.out.println("Stock initialized");
        System.out.println("Application ready.");
    }
}
