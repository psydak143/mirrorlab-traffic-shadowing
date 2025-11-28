package com.mirrorlab.demo.repo;

import com.mirrorlab.demo.model.Product;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.stream.Collectors;

@Repository
public class ProductRepository {
    private final Map<String, Product> products;

    public ProductRepository() {
        Map<String, Product> m = new LinkedHashMap<>();
        m.put("p-100", new Product("p-100", "NVMe SSD 1TB", "storage", 6900));
        m.put("p-101", new Product("p-101", "NVMe SSD 2TB", "storage", 11900));
        m.put("p-102", new Product("p-102", "DDR5 RAM 16GB", "memory", 5200));
        m.put("p-103", new Product("p-103", "DDR5 RAM 32GB", "memory", 9800));
        m.put("p-104", new Product("p-104", "USB-C Hub", "accessories", 2900));
        m.put("p-105", new Product("p-105", "Mechanical Keyboard", "peripherals", 7900));
        m.put("p-106", new Product("p-106", "1080p Webcam", "peripherals", 3400));
        m.put("p-107", new Product("p-107", "27\" 144Hz Monitor", "display", 22900));
        m.put("p-108", new Product("p-108", "Wireless Mouse", "peripherals", 1900));
        m.put("p-109", new Product("p-109", "External SSD 1TB", "storage", 8900));
        products = Collections.unmodifiableMap(m);
    }

    public Optional<Product> findById(String id) {
        return Optional.ofNullable(products.get(id));
    }

    public List<Product> search(String query) {
        String q = Optional.ofNullable(query).orElse("").trim().toLowerCase(Locale.ROOT);
        if (q.isEmpty()) {
            return new ArrayList<>(products.values());
        }
        return products.values().stream()
                .filter(p -> p.name().toLowerCase(Locale.ROOT).contains(q)
                        || p.category().toLowerCase(Locale.ROOT).contains(q)
                        || p.id().toLowerCase(Locale.ROOT).contains(q))
                .collect(Collectors.toList());
    }

    public int total(List<String> ids) {
        if (ids == null) return 0;
        return ids.stream()
                .map(this::findById)
                .filter(Optional::isPresent)
                .mapToInt(opt -> opt.get().priceCents())
                .sum();
    }
}
