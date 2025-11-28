package com.mirrorlab.demo.model;

public record CheckoutResponse(String orderId, int totalCents, long timestamp) { }
