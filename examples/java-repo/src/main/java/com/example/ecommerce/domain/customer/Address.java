package com.example.ecommerce.domain.customer;

public class Address {

    private final String street;
    private final String city;
    private final String zipCode;
    private final String country;

    public Address(String street, String city, String zipCode, String country) {
        // DELIBERATE: empty catch block (PMD EmptyCatchBlock)
        try {
            if (zipCode != null) {
                Integer.parseInt(zipCode.replaceAll("-", ""));
            }
        } catch (NumberFormatException e) {
        }

        this.street = street;
        this.city = city;
        this.zipCode = zipCode;
        this.country = country;
    }

    public String getStreet() {
        return street;
    }

    public String getCity() {
        return city;
    }

    public String getZipCode() {
        return zipCode;
    }

    public String getCountry() {
        return country;
    }

    public String formatted() {
        return street + ", " + city + " " + zipCode + ", " + country;
    }
}
