
## Demo flow

Suggested flow for a recording or interview demo:

1. Start everything:

   ```bash
   cd mirrorlab-traffic-shadowing/mirrorlab
   docker compose up --build
   ```

2. Open Grafana → `MirrorLab` → `MirrorLab – Shadowing Overview`.

3. **Healthy candidate run**:

    - Ensure v2 chaos is off or small in `compose.yaml`:

      ```text
      CHAOS_LATENCY_MS: 0
      CHAOS_JITTER_MS: 0
      CHAOS_ERROR_RATE: 0.0
      ```

    - Run k6 for a few minutes.
    - Show:
        - RPS baseline vs candidate both > 0.
        - p99 latencies close to each other.
        - Error rates ~0 for both.
        - `mirror_enabled` = 1.
        - `mirror_aborts_total` = 0.

4. **Introduce regression in candidate**:

    - Increase chaos for v2:

      ```text
      CHAOS_LATENCY_MS: 200
      CHAOS_JITTER_MS: 100
      CHAOS_ERROR_RATE: 0.02
      ```

    - Restart Compose and re-run k6.
    - Show:
        - Candidate p99 significantly worse than baseline.
        - Candidate error rate rising.
        - Diff mismatches appearing.
        - Guardrail eventually triggers:
            - `mirror_enabled` drops to 0.
            - `mirror_aborts_total` increments.
            - RPS for `target="candidate"` trends down to 0.

5. **Recovery**:

    - Fix candidate chaos settings (back to healthy).
    - Wait for cooldown or manually re-enable mirroring:

      ```bash
      curl -X POST http://localhost:8080/control/enable
      ```

    - Show:
        - `mirror_enabled` back to 1.
        - Candidate RPS re-appearing.
        - Latencies and errors converging again.

---