# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, we appreciate any information about it so that we can fix the problem as soon as possible.

Please report it privately. We prefer reports as GitHub security advisories: open one through "Report a vulnerability" in the [Security tab](https://github.com/teamhanko/hanko/security/advisories/new) of this repository. If you don't have a GitHub account, email us at security@hanko.io instead.

- We will keep your report confidential and will not share your personal information with any third party without your consent
- We will keep you informed of the progress until the problem is resolved

Before you submit, please read the guidelines below. Reports that don't follow them take much longer to triage. We may close a report without further investigation if we cannot reproduce the problem with the steps it gives, or if its claims don't match the code.

## Report Guidelines

### One problem per report

Each report must describe exactly one vulnerability. If you found several problems that can each be exploited and reproduced on their own, submit one report per problem, even if they are in the same area of the code. Only describe several problems in one report if they can only be exploited together.

### Reproduce it against a running instance

A report must show that the vulnerability can be exploited in a running Hanko instance, not only that the code looks wrong.

- State the exact version or commit you tested against. Use the latest release or the current `main` branch.
- State the configuration you used, and point out every setting that differs from the default.
- Give step-by-step instructions to reproduce the problem: the exact HTTP requests (method, path, headers, body) and the responses you received.
- Do not demonstrate the problem by modifying Hanko's code, mocking parts of it, or calling internal functions directly. If the problem cannot be triggered through the API or the user-facing flows, it is most likely not exploitable.

### Describe the attacker

State plainly:

- Who the attacker is, e.g. an unauthenticated visitor, a logged-in user, or a user of another tenant.
- What the attacker needs to have or control before the attack, e.g. an account, a token, a domain, or an identity provider.
- What the attacker gains that they could not get otherwise.

### Show the impact, don't extrapolate it

- Describe the impact you demonstrated. Do not chain hypothetical steps ("this could lead to ..., which could allow ...") unless you demonstrated each of them.
- If you include a severity or CVSS score, explain how you arrived at each metric. Do not rate a problem higher than what you demonstrated.

### Reference the code precisely

- If you point to code, link to the exact file and line at a specific commit.
- Only state how the code behaves after reading it. Do not infer behavior from function names, comments, or search results.

### Use Hanko's terminology

Name things the way Hanko names them: use the names of endpoints, configuration keys, flows, and other concepts as they appear in Hanko's code, API, and documentation. Do not make up your own names or labels for them. If you need to refer to something Hanko has no name for, describe it plainly instead of inventing a term. Every unfamiliar term in a report is something we have to translate back before we can evaluate it.

The same applies to labels from your own setup, e.g. names for your test environment or IDs for your test cases. Avoid them where you can. If you use one, explain it where it first appears, and only refer to things the report has already introduced.

### Format the report properly

Reports are read as Markdown, e.g. in a GitHub security advisory. Do not paste output from a terminal, an AI tool, or an agent without formatting it:

- Do not hard-wrap or cut off lines. Each paragraph is one line; requests, responses, logs, and code are complete.
- Put code, HTTP requests and responses, logs, and command output in fenced code blocks and declare their language, e.g. ` ```http `, ` ```go `, or ` ```shell `.
- Do not attach files to a GitHub security advisory or its comments. Attached files are often not accessible to us. Put their content into the advisory or a comment as fenced code blocks instead.
- Escape or fence anything that Markdown would otherwise render, e.g. HTML tags, `<placeholders>`, `*`, or `_`.
- Check the preview before you submit.

### Get to the point

Write plainly and directly. Every sentence should tell us something about the problem that the rest of the report doesn't already say.

- Start with a summary that states what is wrong, where, and what an attacker gains.
- Do not repeat the same point in different words, e.g. in a summary, an overview, and a conclusion.
- Leave out general explanations of the vulnerability class, filler, and boilerplate.
- A suggested fix is welcome but optional.

### Out of scope

The following are not vulnerabilities in Hanko:

- Access to the Admin API or the Management API. Neither has authentication of its own; both must be protected by the operator. Anyone who can reach them is trusted.
- Anything that requires control over Hanko's configuration, its database, or the host it runs on.
- Problems that only occur with a configuration that contradicts our documentation.
- Vulnerabilities in dependencies, unless you show that the vulnerable code can be reached and exploited through Hanko.

## AI-Assisted Reports

You may use AI tools to find a vulnerability or to write up a report. You remain responsible for the report:

- Verify every claim in the report yourself, and reproduce the problem yourself as described above.
- Before you let an AI assistant write a report, make it read this file and follow the guidelines above.
