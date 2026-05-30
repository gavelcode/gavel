package com.example.ecommerce.application.order;

import com.example.ecommerce.domain.customer.Customer;
import com.example.ecommerce.domain.customer.CustomerRepository;
import com.example.ecommerce.domain.inventory.InventoryRepository;
import com.example.ecommerce.domain.inventory.Stock;
import com.example.ecommerce.domain.order.Money;
import com.example.ecommerce.domain.order.Order;
import com.example.ecommerce.domain.order.OrderLine;
import com.example.ecommerce.domain.order.OrderRepository;
import com.example.ecommerce.domain.product.Product;
import com.example.ecommerce.domain.product.ProductRepository;
import java.util.List;
import java.util.Optional;

public class PlaceOrderHandler {

    private final OrderRepository orderRepo;
    private final CustomerRepository customerRepo;
    private final ProductRepository productRepo;
    private final InventoryRepository inventoryRepo;

    public PlaceOrderHandler(
            OrderRepository orderRepo,
            CustomerRepository customerRepo,
            ProductRepository productRepo,
            InventoryRepository inventoryRepo) {
        this.orderRepo = orderRepo;
        this.customerRepo = customerRepo;
        this.productRepo = productRepo;
        this.inventoryRepo = inventoryRepo;
    }

    // DELIBERATE: long method with high cyclomatic complexity (PMD CyclomaticComplexity)
    public OrderDto execute(long customerId, List<OrderItemRequest> items) {
        Optional<Customer> customerOpt = customerRepo.findById(customerId);
        if (!customerOpt.isPresent()) {
            throw new IllegalArgumentException("customer not found: " + customerId);
        }
        Customer customer = customerOpt.get();

        if (items == null || items.isEmpty()) {
            throw new IllegalArgumentException("order must have at least one item");
        }

        long orderId = orderRepo.nextId();
        Order order = new Order(orderId, customer.getId());

        for (OrderItemRequest item : items) {
            if (item.quantity <= 0) {
                throw new IllegalArgumentException("quantity must be positive for product " + item.productId);
            }

            Optional<Product> productOpt = productRepo.findById(item.productId);
            if (!productOpt.isPresent()) {
                throw new IllegalArgumentException("product not found: " + item.productId);
            }
            Product product = productOpt.get();

            Optional<Stock> stockOpt = inventoryRepo.findByProductId(item.productId);
            if (!stockOpt.isPresent()) {
                throw new IllegalArgumentException("no stock for product: " + item.productId);
            }
            Stock stock = stockOpt.get();

            if (!stock.hasEnough(item.quantity)) {
                throw new IllegalStateException(
                    "insufficient stock for " + product.getName()
                        + ": available=" + stock.getQuantity()
                        + ", requested=" + item.quantity);
            }

            stock.reserve(item.quantity);
            inventoryRepo.save(stock);

            OrderLine line = new OrderLine(
                product.getId(),
                product.getName(),
                item.quantity,
                product.getPrice());
            order.addLine(line);
        }

        Money total = order.total();
        if (total.isZero()) {
            throw new IllegalStateException("order total cannot be zero");
        }

        order.confirm();
        orderRepo.save(order);

        return OrderDto.fromDomain(order);
    }

    public static class OrderItemRequest {
        public final long productId;
        public final int quantity;

        public OrderItemRequest(long productId, int quantity) {
            this.productId = productId;
            this.quantity = quantity;
        }
    }
}
