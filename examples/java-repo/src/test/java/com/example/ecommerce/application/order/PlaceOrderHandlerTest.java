package com.example.ecommerce.application.order;

import static org.junit.Assert.*;

import com.example.ecommerce.domain.customer.Customer;
import com.example.ecommerce.domain.customer.CustomerRepository;
import com.example.ecommerce.domain.inventory.InventoryRepository;
import com.example.ecommerce.domain.inventory.Stock;
import com.example.ecommerce.domain.order.Money;
import com.example.ecommerce.domain.order.Order;
import com.example.ecommerce.domain.order.OrderRepository;
import com.example.ecommerce.domain.product.Product;
import com.example.ecommerce.domain.product.ProductRepository;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicLong;
import org.junit.Before;
import org.junit.Test;

public class PlaceOrderHandlerTest {

    private PlaceOrderHandler handler;
    private Map<Long, Customer> customers;
    private Map<Long, Product> products;
    private Map<Long, Stock> stocks;
    private Map<Long, Order> orders;
    private AtomicLong orderIdSeq;

    @Before
    public void setUp() {
        customers = new HashMap<>();
        products = new HashMap<>();
        stocks = new HashMap<>();
        orders = new HashMap<>();
        orderIdSeq = new AtomicLong(1);

        customers.put(1L, new Customer(1, "Alice", "alice@example.com"));
        products.put(1L, new Product(1, "Laptop", "Gaming laptop", Money.of(999.99, "USD")));
        products.put(2L, new Product(2, "Mouse", "Wireless mouse", Money.of(29.99, "USD")));
        stocks.put(1L, new Stock(1, 10));
        stocks.put(2L, new Stock(2, 50));

        OrderRepository orderRepo = new OrderRepository() {
            @Override
            public void save(Order order) {
                orders.put(order.getId(), order);
            }

            @Override
            public Optional<Order> findById(long id) {
                return Optional.ofNullable(orders.get(id));
            }

            @Override
            public long nextId() {
                return orderIdSeq.getAndIncrement();
            }
        };

        CustomerRepository customerRepo = new CustomerRepository() {
            @Override
            public void save(Customer c) {
                customers.put(c.getId(), c);
            }

            @Override
            public Optional<Customer> findById(long id) {
                return Optional.ofNullable(customers.get(id));
            }
        };

        ProductRepository productRepo = id -> Optional.ofNullable(products.get(id));

        InventoryRepository inventoryRepo = new InventoryRepository() {
            @Override
            public void save(Stock s) {
                stocks.put(s.getProductId(), s);
            }

            @Override
            public Optional<Stock> findByProductId(long productId) {
                return Optional.ofNullable(stocks.get(productId));
            }
        };

        handler = new PlaceOrderHandler(orderRepo, customerRepo, productRepo, inventoryRepo);
    }

    @Test
    public void shouldPlaceOrderSuccessfully() {
        List<PlaceOrderHandler.OrderItemRequest> items = List.of(
            new PlaceOrderHandler.OrderItemRequest(1, 2),
            new PlaceOrderHandler.OrderItemRequest(2, 3));

        OrderDto result = handler.execute(1, items);

        assertNotNull(result);
        assertEquals(1, result.id);
        assertEquals(1, result.customerId);
        assertEquals("CONFIRMED", result.status);
        assertEquals(2, result.lines.size());
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectUnknownCustomer() {
        handler.execute(999, List.of(new PlaceOrderHandler.OrderItemRequest(1, 1)));
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectEmptyItems() {
        handler.execute(1, List.of());
    }

    @Test(expected = IllegalStateException.class)
    public void shouldRejectInsufficientStock() {
        handler.execute(1, List.of(new PlaceOrderHandler.OrderItemRequest(1, 100)));
    }
}
