package com.example.ecommerce.platform.config;

public class AppConfig {

    // DELIBERATE: mutable static field (SpotBugs MS_SHOULD_BE_FINAL)
    public static String APP_NAME = "ecommerce";
    public static String APP_VERSION = "1.0.0";

    private final String environment;

    public AppConfig(String environment) {
        if (environment == null || environment.isBlank()) {
            this.environment = "development";
        } else {
            this.environment = environment;
        }
    }

    public String getEnvironment() {
        return environment;
    }

    public boolean isProduction() {
        return "production".equals(environment);
    }

    public String getAppName() {
        return APP_NAME;
    }

    public String getAppVersion() {
        return APP_VERSION;
    }
}
