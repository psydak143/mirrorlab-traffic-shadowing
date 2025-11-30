## Future improvements

Ideas for next iterations:

- Kubernetes manifests or a Helm chart.
- Canary step-ups (automatically increasing `MIRROR_FRACTION` if candidate is healthy).
- Schema-aware or OpenAPI-based diff and PII redaction.
- OpenTelemetry traces and logs correlated per request.
- CI/CD:
    - GitHub Actions for tests, linters, vulnerability scans.
    - Automated image builds and pushes.

---