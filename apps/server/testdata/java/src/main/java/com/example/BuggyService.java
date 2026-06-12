package com.example;

import java.util.ArrayList;
import java.util.List;

public final class BuggyService {

    public void deadException(String input) {
        if (input == null) {
            new IllegalArgumentException("input must not be null");
        }
        System.out.println(input);
    }

    public String nullDereference() {
        String value = null;
        return value.trim();
    }

    public boolean badEquals(String a, String b) {
        return a == b;
    }

    public List<String> modifyWhileIterating(List<String> items) {
        List<String> result = new ArrayList<>(items);
        for (String item : result) {
            if (item.isEmpty()) {
                result.remove(item);
            }
        }
        return result;
    }

    @SuppressWarnings("all")
    public void suppressedWarning() {
        int unused = 42;
    }
}
