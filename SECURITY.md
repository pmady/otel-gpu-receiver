# Security Policy

## Reporting a Vulnerability

Don't open a public GitHub issue for security bugs. Report them privately:

1. [GitHub private vulnerability reporting](https://github.com/pmady/otel-gpu-receiver/security/advisories/new)
2. Or email **pavan4devops@gmail.com**

Include: what the bug is, how to reproduce it, and the impact. I'll acknowledge within 48 hours and coordinate a fix + disclosure timeline.

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅        |

## Deployment Notes

- Run the Collector with least-privilege RBAC
- Use network policies to limit Collector egress
- Don't expose NVML metrics endpoints without auth
