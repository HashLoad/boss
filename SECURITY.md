# Security Policy

## Supported Versions

We actively support and provide security patches for the following versions of Boss:

| Version | Supported          |
| ------- | ------------------ |
| latest  | ✅ Yes             |
| < 2.0   | ❌ No              |

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

To report a security issue, please use one of the following methods:

1. **GitHub Private Vulnerability Reporting** (preferred):
   Navigate to the [Security Advisories](https://github.com/HashLoad/boss/security/advisories/new) page
   and submit a private advisory.

2. **Email**: Send details to `security@hashload.com` with the subject line:
   `[SECURITY] Boss - <brief description>`

### What to include

- Description of the vulnerability and its potential impact
- Steps to reproduce or proof-of-concept
- Affected version(s) and environment (OS, Delphi version)
- Any suggested mitigation or fix

### Response Timeline

| Stage             | Target SLA        |
| ----------------- | ----------------- |
| Acknowledgement   | ≤ 3 business days |
| Initial triage    | ≤ 7 business days |
| Fix / Advisory    | ≤ 90 days         |

We will keep you informed throughout the process and credit you in the release notes
(unless you prefer to remain anonymous).

## Scope

This policy covers the **Boss CLI binary** (`boss.exe` / `boss`) and the Go source code
in this repository. It does **not** cover:

- Third-party packages installed via `boss install` (report those to their respective maintainers)
- The PubPascal portal (report to `security@pubpascal.dev`)

## CRA Compliance

Boss ships a machine-readable **Software Bill of Materials (SBOM)** with every release
and maintains this vulnerability-disclosure policy in accordance with the
[EU Cyber Resilience Act (CRA)](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32024R2847)
Article 14 (active vulnerability management) and Annex I Part II (secure development).

The SBOM is published as `sbom.cdx.json` (CycloneDX 1.6) at each
[GitHub release](https://github.com/HashLoad/boss/releases).
