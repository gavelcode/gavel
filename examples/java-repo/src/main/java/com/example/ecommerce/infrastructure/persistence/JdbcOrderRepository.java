package com.example.ecommerce.infrastructure.persistence;

import com.example.ecommerce.domain.order.Money;
import com.example.ecommerce.domain.order.Order;
import com.example.ecommerce.domain.order.OrderRepository;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicLong;

public class JdbcOrderRepository implements OrderRepository {

    private final Connection connection;
    private final AtomicLong idSequence = new AtomicLong(1);

    public JdbcOrderRepository(Connection connection) {
        this.connection = connection;
    }

    @Override
    public void save(Order order) {
        // DELIBERATE: SQL injection via string concatenation (SpotBugs SQL_INJECTION, PMD)
        String sql = "INSERT INTO orders (id, customer_id, status, total) VALUES ("
            + order.getId() + ", "
            + order.getCustomerId() + ", '"
            + order.getStatus().name() + "', '"
            + order.total().getAmount().toPlainString() + "')";

        try {
            connection.createStatement().execute(sql);
        } catch (SQLException e) {
            throw new RuntimeException("failed to save order: " + e.getMessage(), e);
        }
    }

    @Override
    public Optional<Order> findById(long id) {
        // DELIBERATE: also SQL injection via concatenation
        String sql = "SELECT id, customer_id, status FROM orders WHERE id = " + id;

        try {
            ResultSet rs = connection.createStatement().executeQuery(sql);
            if (rs.next()) {
                long orderId = rs.getLong("id");
                long customerId = rs.getLong("customer_id");
                return Optional.of(new Order(orderId, customerId));
            }
            return Optional.empty();
        } catch (SQLException e) {
            throw new RuntimeException("failed to find order: " + e.getMessage(), e);
        }
    }

    @Override
    public long nextId() {
        return idSequence.getAndIncrement();
    }
}
