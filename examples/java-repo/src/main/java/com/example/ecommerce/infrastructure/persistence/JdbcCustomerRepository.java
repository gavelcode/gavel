package com.example.ecommerce.infrastructure.persistence;

import com.example.ecommerce.domain.customer.Customer;
import com.example.ecommerce.domain.customer.CustomerRepository;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.Optional;

public class JdbcCustomerRepository implements CustomerRepository {

    private final Connection connection;

    public JdbcCustomerRepository(Connection connection) {
        this.connection = connection;
    }

    @Override
    public void save(Customer customer) {
        String sql = "INSERT INTO customers (id, name, email) VALUES (?, ?, ?)";
        try {
            PreparedStatement stmt = connection.prepareStatement(sql);
            stmt.setLong(1, customer.getId());
            stmt.setString(2, customer.getName());
            stmt.setString(3, customer.getEmail());
            stmt.executeUpdate();
            // DELIBERATE: PreparedStatement not closed (SpotBugs OBL_UNSATISFIED_OBLIGATION)
        } catch (SQLException e) {
            throw new RuntimeException("failed to save customer: " + e.getMessage(), e);
        }
    }

    @Override
    public Optional<Customer> findById(long id) {
        String sql = "SELECT id, name, email FROM customers WHERE id = ?";
        try {
            PreparedStatement stmt = connection.prepareStatement(sql);
            stmt.setLong(1, id);
            ResultSet rs = stmt.executeQuery();
            // DELIBERATE: ResultSet and PreparedStatement not closed
            if (rs.next()) {
                return Optional.of(new Customer(
                    rs.getLong("id"),
                    rs.getString("name"),
                    rs.getString("email")));
            }
            return Optional.empty();
        } catch (SQLException e) {
            throw new RuntimeException("failed to find customer: " + e.getMessage(), e);
        }
    }
}
