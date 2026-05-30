package com.example.ecommerce.application.order;

import com.example.ecommerce.domain.order.Order;
import com.example.ecommerce.domain.order.OrderLine;
import java.util.ArrayList;
import java.util.List;

public class OrderDto {

    public final long id;
    public final long customerId;
    public final String status;
    public final String total;
    public final List<LineDto> lines;

    public OrderDto(long id, long customerId, String status, String total, List<LineDto> lines) {
        this.id = id;
        this.customerId = customerId;
        this.status = status;
        this.total = total;
        this.lines = lines;
    }

    public static OrderDto fromDomain(Order order) {
        // DELIBERATE: unused variable (PMD UnusedLocalVariable)
        int lineCount = order.getLines().size();

        List<LineDto> lineDtos = new ArrayList<>();
        for (OrderLine line : order.getLines()) {
            lineDtos.add(new LineDto(
                line.getProductId(),
                line.getProductName(),
                line.getQuantity(),
                line.getUnitPrice().toString()));
        }
        return new OrderDto(
            order.getId(),
            order.getCustomerId(),
            order.getStatus().name(),
            order.total().toString(),
            lineDtos);
    }

    public static class LineDto {
        public final long productId;
        public final String productName;
        public final int quantity;
        public final String unitPrice;

        public LineDto(long productId, String productName, int quantity, String unitPrice) {
            this.productId = productId;
            this.productName = productName;
            this.quantity = quantity;
            this.unitPrice = unitPrice;
        }
    }
}
