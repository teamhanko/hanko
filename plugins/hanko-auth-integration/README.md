# Hanko Agent Plugin

This plugin packages a focused agent workflow for Hanko maintainers and users. It is designed to be useful in Codex, Claude Code, Claude Cowork, Copilot-style coworkers, and other `SKILL.md`-compatible harnesses.

The plugin does not add a runtime dependency to Hanko. It gives agents a precise operating procedure, expected outputs, and plugin evals so maintainers can decide whether agent-produced work is good enough to accept.

## What It Includes

- Codex and Claude plugin manifests.
- A Hanko-specific skill at `skills/hanko-auth-integration/SKILL.md`.
- Plugin eval cases in `evals/hanko-auth-integration/cases.jsonl`.
- Privacy-safe measurement guidance for teams that want production plugin metrics.

## Primary Workflows

- Framework integration plan.
- Passkey flow review.
- Session and redirect audit.
- Example-app regression pack.

## Eval Cases

- `nextjs-passkey`: Plan a Hanko passkey integration for a Next.js app.
- `redirect-audit`: Audit an auth configuration where production sign-in redirects to localhost.
- `session-regression`: Create eval cases for sign-up, sign-in, sign-out, and session refresh.

## Install In An Agent Harness

Use this plugin directory directly from the repository when your harness supports local or Git-backed plugin sources. The plugin root is:

```text
plugins/hanko-auth-integration
```

For Telvine-backed distribution and metrics:

```bash
npm i -g telvine
telvine login
telvine publish ./plugins/hanko-auth-integration
telvine plugins metrics
```

## Telemetry Boundary

The plugin should only record metadata about plugin execution and eval outcomes. Do not record prompts, source files, request bodies, connector payloads, credentials, model outputs, or production user data.
