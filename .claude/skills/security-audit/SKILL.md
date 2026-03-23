---
name: security-audit
description: >
  Comprehensive security audit for full-stack web applications (Go backend, React/TypeScript frontend, Docker, Nginx).
  Use this skill whenever the user asks to audit, review security, check for vulnerabilities, harden, or pen-test
  their codebase — even if they just say "is this secure?" or "check for security issues." Also triggers from the
  dev-pipeline skill as a pre-PR gate.
---

# Security Audit

You are performing a thorough, multi-phase security audit of a codebase. This is not a cursory scan — you are acting as a senior application security engineer doing a manual code review supplemented by automated tooling.

The audit produces a **Security Report** with findings rated by severity, plus a pass/fail gate verdict.

## Audit Phases

Work through these phases in order. For each phase, log what you checked and what you found.

### Phase 1: Dependency Vulnerabilities

Scan every dependency manifest for known CVEs and outdated packages.

**Go backend:**
```bash
# Check for known vulnerabilities in Go modules
cd backend && go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...
# List outdated modules
go list -m -u all
```

**Frontend (npm):**
```bash
cd frontend && npm audit --audit-level=moderate
```

If govulncheck or npm audit finds high/critical issues, flag them as CRITICAL findings.

### Phase 2: Static Analysis — Backend (Go)

Review Go source code for these vulnerability classes:

- **SQL injection**: String concatenation or fmt.Sprintf in SQL queries instead of parameterized queries ($1, $2). Check every file in internal/store/ and any file that touches database/sql.
- **Command injection**: Search for os/exec, os.Command, or any shell invocation with user-controlled input.
- **SSRF**: HTTP client calls where the URL is derived from user input. Focus on internal/scraper/.
- **Path traversal**: os.Open, filepath.Join, or file operations where the path includes user input.
- **Information leakage**: Error messages that expose stack traces, internal paths, or database details to the client.
- **Race conditions**: Shared mutable state accessed without synchronization. Goroutines writing to maps or slices.
- **Hardcoded secrets**: Grep for API keys, passwords, tokens in source files (not just .env).
- **Auth/authz flaws**: Endpoints missing authentication checks. Spotify token validation.

Run if available:
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest && gosec ./...
```

### Phase 3: Static Analysis — Frontend (TypeScript/React)

Review frontend source for these vulnerability classes:

- **XSS vectors**: Unsafe DOM manipulation, raw HTML injection, dynamic code evaluation. Search for patterns that bypass React's built-in escaping.
- **Open redirects**: Redirect targets sourced from URL params or user input.
- **Sensitive data in client state**: How are tokens stored — localStorage, sessionStorage, or memory? Check Spotify token handling.
- **CORS misconfiguration**: Requests to untrusted origins.
- **Dependency supply chain**: Suspicious or unmaintained packages in package.json.

Run the existing linter:
```bash
cd frontend && npm run lint
```

### Phase 4: Infrastructure & Configuration

- **Docker**: Running as root? Unnecessary capabilities? Minimal base image? Secrets baked into layers?
- **Nginx**: Missing security headers? TLS misconfiguration? Open proxy risks? Rate limiting gaps? Verify HSTS, CSP, X-Frame-Options, X-Content-Type-Options.
- **Docker Compose**: Exposed ports that should be internal-only? Missing health checks? Missing resource limits?
- **Environment variables**: Is .env in .gitignore? Does .env.example contain real values?

### Phase 5: Authentication & Session Security

- **OAuth2 PKCE flow**: Code verifier entropy, secure transmission, server-side token exchange.
- **Token storage**: Are tokens exposed to script-based attacks?
- **Token refresh**: Is the refresh flow secure? Can tokens be replayed?
- **Session fixation**: Can an attacker pre-set a session token before auth?

### Phase 6: API Security

- **Input validation**: Are all API inputs validated and sanitized? Query params, path params, request bodies.
- **Rate limiting**: Applied to all sensitive endpoints? Check for bypass vectors.
- **Error handling**: Do error responses leak internal details?
- **CORS policy**: Is FRONTEND_URL the only allowed origin? Is wildcard used anywhere?

### Phase 7: Secrets & Credential Scanning

```bash
# Search for potential hardcoded secrets across the entire repo
grep -rn --include="*.go" --include="*.ts" --include="*.tsx" --include="*.js" --include="*.yml" --include="*.yaml" --include="*.json" \
  -E '(password|secret|token|api_key|apikey|private_key|aws_|AKIA)' . \
  | grep -v node_modules | grep -v '.env.example' | grep -v go.sum
```

Also check git history for accidentally committed secrets:
```bash
git log --all --diff-filter=A -- '*.env' '*.pem' '*.key' '*credentials*'
```

## Security Report Format

After completing all phases, produce a report:

```
# Security Audit Report

Date: [date]
Scope: [what was audited]
Verdict: PASS / FAIL (fail if any HIGH or CRITICAL findings)

## Summary
- Critical: [count]
- High: [count]
- Medium: [count]
- Low: [count]
- Informational: [count]

## Findings

### [SEVERITY] Finding Title
- Phase: [which audit phase]
- Location: file:line
- Description: What the issue is
- Impact: What an attacker could do
- Recommendation: How to fix it
- Evidence: The specific code or config

[repeat for each finding]

## Checklist
- [ ] Dependencies scanned
- [ ] Go static analysis complete
- [ ] Frontend static analysis complete
- [ ] Infrastructure reviewed
- [ ] Auth flow reviewed
- [ ] API security reviewed
- [ ] Secrets scan complete
```

## Gate Verdict

The audit passes only if there are zero CRITICAL or HIGH findings. Medium and below are acceptable for the PR to proceed, but should be tracked as follow-up issues.

If the audit fails, list the specific findings that must be fixed before the PR can be opened.
