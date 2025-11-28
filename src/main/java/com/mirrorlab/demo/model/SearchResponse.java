package com.mirrorlab.demo.model;

import java.util.List;

public record SearchResponse(String query, int count, List<Product> items, long timestamp) { }
