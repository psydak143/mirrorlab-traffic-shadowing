package com.mirrorlab.demo.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.mirrorlab.demo.chaos.ChaosProperties;
import com.mirrorlab.demo.model.*;
import com.mirrorlab.demo.repo.ProductRepository;
import com.mirrorlab.demo.util.DeterministicIds;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;

import java.util.List;
import java.util.Locale;
import java.util.concurrent.ThreadLocalRandom;

@RestController
@RequestMapping(path = "/api", produces = MediaType.APPLICATION_JSON_VALUE)
public class ProductController {
    private final ProductRepository repo;
    private final ChaosProperties chaos;
    // Create our own ObjectMapper instead of injecting it
    private final ObjectMapper mapper = new ObjectMapper();

    public ProductController(ProductRepository repo, ChaosProperties chaos) {
        this.repo = repo;
        this.chaos = chaos;
    }

    @GetMapping("/search")
    public SearchResponse search(@RequestParam(name = "q", required = false) String q) {
        maybeInjectError("GET /search");
        List<Product> items = repo.search(q);
        return new SearchResponse(q == null ? "" : q, items.size(), items, System.currentTimeMillis());
    }

    @GetMapping("/product/{id}")
    public Product product(@PathVariable String id) {
        maybeInjectError("GET /product/{id}");
        return repo.findById(id).orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "no such product"));
    }

    @PostMapping(path = "/checkout", consumes = MediaType.APPLICATION_JSON_VALUE)
    public CheckoutResponse checkout(@RequestBody CheckoutRequest req) {
        maybeInjectError("POST /checkout");
        int total = repo.total(req.productIds());
        String normalized = toCanonicalJson(req).toLowerCase(Locale.ROOT) + "|" + total;
        String orderId = DeterministicIds.sha256Hex(normalized).substring(0, 16); // short & deterministic
        return new CheckoutResponse(orderId, total, System.currentTimeMillis());
    }

    private void maybeInjectError(String route) {
        double rate = chaos.getErrorRate();
        if (rate <= 0.0) return;
        if (ThreadLocalRandom.current().nextDouble() < rate) {
            throw new ResponseStatusException(HttpStatus.INTERNAL_SERVER_ERROR, "chaos: injected error for " + route);
        }
    }

    private String toCanonicalJson(Object o) {
        try {
            return mapper.writer().withDefaultPrettyPrinter().writeValueAsString(o);
        } catch (JsonProcessingException e) {
            return o.toString();
        }
    }
}
