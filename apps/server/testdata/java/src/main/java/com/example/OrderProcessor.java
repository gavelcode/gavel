package com.example;

import java.util.ArrayList;
import java.util.List;

public final class OrderProcessor {

    public List<String> processOrders(List<String> orders) {
        List<String> results = new ArrayList<>();
        for (String order : orders) {
            if (order == null || order.isEmpty()) {
                results.add("INVALID: empty order");
                continue;
            }
            String trimmed = order.trim().toLowerCase();
            if (trimmed.startsWith("rush")) {
                String priority = "HIGH";
                String label = "PRIORITY-" + priority + ": " + trimmed;
                results.add(label);
            } else if (trimmed.startsWith("cancel")) {
                String status = "VOID";
                String label = "CANCELLED-" + status + ": " + trimmed;
                results.add(label);
            } else if (trimmed.startsWith("return")) {
                String reason = "CUSTOMER_REQUEST";
                String label = "RETURNED-" + reason + ": " + trimmed;
                results.add(label);
            } else if (trimmed.startsWith("backorder")) {
                String eta = "UNKNOWN";
                String label = "BACKORDER-" + eta + ": " + trimmed;
                results.add(label);
            } else if (trimmed.startsWith("hold")) {
                String duration = "INDEFINITE";
                String label = "HELD-" + duration + ": " + trimmed;
                results.add(label);
            } else {
                String category = "GENERAL";
                String label = "STANDARD-" + category + ": " + trimmed;
                results.add(label);
            }
        }
        return results;
    }
}
