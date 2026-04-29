# Security Policy

`reclaim` deletes files. We take security and correctness reports seriously.

## Reporting a vulnerability

**Do not open a public GitHub issue** for vulnerabilities. Instead:

1. Open a private security advisory:
   <https://github.com/ImadRashid/reclaim/security/advisories/new>
2. Or email the maintainer at imadrashid789@gmail.com with the subject
   "[reclaim] security report".

Include:

- A description of the issue
- The version of `reclaim` affected (`reclaim --version`)
- Steps to reproduce
- The impact (what gets deleted, what privilege is required, etc.)

You will receive an acknowledgement within 72 hours and a status update at
least every 7 days until the issue is resolved.

## Scope

In scope:

- Path-traversal or path-injection that lets a rule delete files outside the
  intended target
- Bypasses of `CheckPathSafe` in `internal/engine/safety.go`
- Symlink-following that could escape the intended target
- Privilege escalation (`reclaim` runs as the invoking user; it should never
  prompt for or use `sudo`)
- Code execution from malformed YAML rule files

Out of scope:

- Theoretical issues that require an attacker who already has full
  shell access as the user
- Issues in dependencies that don't affect `reclaim`'s actual behavior
- Race conditions where a user manually creates a `Cargo.toml` mid-scan

## Disclosure

We follow coordinated disclosure. After a fix is released, we'll publish a
GitHub Security Advisory crediting you (unless you prefer to remain anonymous).
