# Security Policy

## Supported versions

| Version   | Supported          |
| --------- | ------------------ |
| 1.x       | Yes                |
| &lt; 1.0  | Pre-1.0 / RC only  |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

1. **Preferred:** Use [GitHub private vulnerability reporting](https://github.com/codethor0/dockercomms/security/advisories/new) if it is enabled for this repository (Settings → Security → Code security and analysis → Private vulnerability reporting).

2. **Alternative:** Email the maintainers with a clear subject line (e.g. `[SECURITY] DockerComms`) and enough detail to reproduce or assess the issue. Do not include exploit code in the first message if it is not necessary.

We aim to acknowledge reports within a few business days and coordinate disclosure once a fix is ready.

## What to include

- Affected component (CLI command, package path, or flow)
- Steps or proof-of-concept sufficient to understand impact
- Version or commit if known (`dockercomms version` output when applicable)
- Your assessment of severity (best effort is fine)

## Scope

In scope for security reports:

- The `dockercomms` CLI and libraries in this repository
- Cryptographic verification, path handling, and registry interaction as implemented here

Out of scope (unless they clearly affect this codebase):

- Third-party registry outages or policy
- Cosign / Sigstore infrastructure (report upstream when appropriate)
- Issues in dependencies without a concrete impact on DockerComms (still welcome as regular issues if reproducible)

## Safe harbor

We appreciate responsible disclosure. If you act in good faith and follow this policy, we will not pursue legal action for accidental, good-faith violations.
