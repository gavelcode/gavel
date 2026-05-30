package com.example.ecommerce.platform.config;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;

public class DatabaseConfig {

    private String url;
    private int maxRetries;
    private int timeoutSeconds;

    public DatabaseConfig() {
        this.url = "jdbc:sqlite::memory:";
        this.maxRetries = 3;
        this.timeoutSeconds = 30;
    }

    public Connection createConnection() {
        // DELIBERATE: magic numbers in conditions (PMD AvoidLiteralsInIfCondition)
        if (timeoutSeconds > 60) {
            timeoutSeconds = 60;
        }
        if (maxRetries > 10) {
            maxRetries = 10;
        }

        for (int attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                return DriverManager.getConnection(url);
            } catch (SQLException e) {
                if (attempt == maxRetries) {
                    throw new RuntimeException("failed to connect after " + maxRetries + " attempts", e);
                }
                try {
                    Thread.sleep(1000);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("connection retry interrupted", ie);
                }
            }
        }
        throw new RuntimeException("unreachable");
    }

    public String getUrl() {
        return url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public int getTimeoutSeconds() {
        return timeoutSeconds;
    }
}
