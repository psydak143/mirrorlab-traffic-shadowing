package com.mirrorlab.demo.chaos;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "chaos")
public class ChaosProperties {
    private int latencyMs = 0;     // base delay per request
    private int jitterMs = 0;      // 0..jitter added to latencyMs
    private double errorRate = 0.0; // 0.02 = 2% errors

    public int getLatencyMs() { return latencyMs; }
    public void setLatencyMs(int latencyMs) { this.latencyMs = latencyMs; }

    public int getJitterMs() { return jitterMs; }
    public void setJitterMs(int jitterMs) { this.jitterMs = jitterMs; }

    public double getErrorRate() { return errorRate; }
    public void setErrorRate(double errorRate) { this.errorRate = errorRate; }
}
