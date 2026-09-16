# Security Policy

This repository is source-visible but not open to unsolicited pull requests (see
[CONTRIBUTING.md](CONTRIBUTING.md)). It is part of a commerce analytics platform that
processes real business and order data for a real Shopify store.

## Reporting a vulnerability

Please report suspected vulnerabilities privately via **GitHub's private vulnerability
reporting** for this repository (Security tab -> "Report a vulnerability") rather than a
public issue. If that isn't available, contact me@terencehegarty.com.

Please include:

- A description of the vulnerability and its potential impact
- Steps to reproduce, or a proof of concept if you have one
- The commit/version you tested against

I'll acknowledge reports within a few days. This is a small, personally-run project —
there's no formal SLA, but security reports get priority over feature work.

## Scope

This repo (`shop_platform`) is a shared library with no network-facing code of its own.
Vulnerabilities here are most likely to matter through the services that import it
(`shop_ingestor`, `shop_analytics`, `shop_notifier`) — feel free to report against
whichever repo you found the issue in.

## What NOT to report here

- Findings that require access to production secrets, credentials, or infrastructure you
  don't have and were not given for testing.
- Automated dependency-scanner output with no demonstrated impact (Dependabot and CodeQL
  already run on every repo in this project — check they haven't already flagged it).

## Secrets handling

No secret, API key, credential, or customer data should ever be committed to this or any
other repo in this platform. If you find one, please report it immediately via the private
channel above rather than opening a public issue — even though it will also be rotated on
our end as soon as we're aware.
