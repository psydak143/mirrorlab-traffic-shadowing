package com.mirrorlab.demo.model;

import java.util.List;

public record CheckoutRequest(List<String> productIds, String email) { }
