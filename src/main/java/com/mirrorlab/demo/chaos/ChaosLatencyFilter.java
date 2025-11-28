package com.mirrorlab.demo.chaos;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.concurrent.ThreadLocalRandom;

/**
 * Adds per-request latency for chaos testing (v2 instance).
 * v1 instance simply sets CHAOS_LATENCY_MS=0, CHAOS_ERROR_RATE=0.0.
 */
@Component
@Order(10)
public class ChaosLatencyFilter extends OncePerRequestFilter {
    private final ChaosProperties props;

    public ChaosLatencyFilter(ChaosProperties props) {
        this.props = props;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response,
                                    FilterChain filterChain) throws ServletException, IOException {
        int base = props.getLatencyMs();
        int jitter = props.getJitterMs();
        if (base > 0 || jitter > 0) {
            int extra = (jitter > 0) ? ThreadLocalRandom.current().nextInt(jitter + 1) : 0;
            long sleep = (long) base + extra;
            try { Thread.sleep(sleep); } catch (InterruptedException ignored) { Thread.currentThread().interrupt(); }
        }
        filterChain.doFilter(request, response);
    }
}
