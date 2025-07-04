<p align="center">
  <img width="300" src="https://user-images.githubusercontent.com/20115649/176922807-fb92327a-15d5-4568-a4e7-78093cea045e.svg?sanitize=true#gh-light-mode-only">
  <img width="300" src="https://user-images.githubusercontent.com/20115649/176922819-61dfb644-529f-4f81-a577-7daa47185300.svg?sanitize=true#gh-dark-mode-only">
</p>

---
[![Test Status](https://github.com/teamhanko/hanko/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/teamhanko/hanko/actions/workflows/codeql-analysis.yml)
[![Build Status](https://github.com/teamhanko/hanko/workflows/Go/badge.svg)](https://github.com/teamhanko/hanko/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/teamhanko/hanko)](https://goreportcard.com/report/github.com/teamhanko/hanko)
[![GoDoc](https://godoc.org/github.com/teamhanko/hanko?status.svg)](https://godoc.org/github.com/teamhanko/hanko)
[![npm (scoped)](https://img.shields.io/npm/v/@teamhanko/hanko-elements?label=hanko-elements)](https://www.npmjs.com/package/@teamhanko/hanko-elements)
[![npm (scoped)](https://img.shields.io/npm/v/@teamhanko/hanko-frontend-sdk?label=hanko-frontend-sdk)](https://www.npmjs.com/package/@teamhanko/hanko-frontend-sdk)

# About Hanko
Hanko is an open source authentication and user management solution that is easy to integrate, framework-agnostic, and built on privacy-first principles like data minimalism and phishing resistance.

- Supports all modern authentication methods: passwords, MFA, passkeys, social logins, and SAML SSO
- Flexible configuration options, including passkey-only, OAuth-only, and user-deletable passwords
- Easy integration with **Hanko Elements** web components
- A robust API that handles all authentication and onboarding flow states, enabling fast, reliable custom frontend implementations
- API-first, lightweight, cloud-native

Available for self-hosting and as a fully managed service on [Hanko Cloud](https://www.hanko.io).

# Features
To follow the development of this project, watch our releases, leave a star, sign up to our [Product News](https://www.hanko.io/updates) or join our [Discord Community](https://www.hanko.io/community). Here's a brief overview of Hanko's current and upcoming features:

| Status | Feature |
|:------:| :--- |
|✅| Email / username identifiers |
|✅| Passwords, passcodes, passkeys |
|✅| Hanko Elements web components |
|✅| OAuth SSO (Sign in with Apple/Google/GitHub and more) |
|✅| i18n & custom translations |
|✅| SAML Enterprise SSO |
|✅| Webhooks |
|✅| Server-side sessions & remote session revocation |
|✅| MFA (TOTP, security keys) |
|✅| Custom OIDC/OAuth connections |
|✅| JS SDK |
|⚙️| Organizations, Roles, Permissions |
| | `<hanko-menu>` web component |
| | iOS, Android, React Native, Flutter SDKs |

Visit our [Roadmap](https://www.hanko.io/roadmap) for more information on upcoming features.

# Contact us
Schedule a demo with the team. Learn how you can built state-of-the-art authentication for your apps effortlessly with Hanko.

<a target="_blank" href="https://cal.com/team/hanko/demo"><img alt="Book us with Cal.com"  src="https://cal.com/book-with-cal-light.svg" /></a>

# Architecture
The main building blocks of the Hanko project are
- [backend](/backend/README.md) - Scalable, robust, and lightweight authentication API for passwords, passkeys, email passcodes, OAuth SSO, user and session management, and JWT issuing
- [hanko-elements](/frontend/elements/README.md) - Web components made for the Hanko API that provide onboarding, login, and user profile functionality and are customizable with CSS
- [hanko-frontend-sdk](/frontend/frontend-sdk/README.md) - A client package for using the Hanko API

The remainder of the repository consists of:
- [quickstart](/quickstart) - A quickstart example app showing off Hanko's login experience and acting as a reference implementation
- [examples](frontend/examples) - Example implementations for a number of frameworks
- docs - The Hanko documentation ([docs.hanko.io](https://docs.hanko.io)) -> Moved to its own repo here: https://github.com/teamhanko/docs

# Getting started
1. Try our hosted [live example](https://example.hanko.io) and our companion page [passkeys.io](https://www.passkeys.io) or use the [quickstart app](/quickstart/README.md) to get a feel for the user experience provided by an application that leverages the Hanko backend API and our custom web component
2. To run the project locally, there are two options available:
   - Bare metal:
      - Head over to the [backend](/backend/README.md) section to learn how to get it up and running for your own project. Use [Hanko Cloud](https://cloud.hanko.io) for a hosted backend.
   - Docker:
     -  If you prefer to use [Docker](https://www.docker.com/) to run the project locally, please visit the [Run the quickstart](./quickstart/README.md#run-the-quickstart) for information on how to run the project. This will create everything, including frontend and backend components. 
        -  If you wish to keep only the backend components, you can modify the [quickstart.yaml](./deploy/docker-compose/quickstart.yaml) to remove the unnecessary services. To make changes to the configuration to meet your needs, modify [config.yaml](./deploy/docker-compose/config.yaml).
3. Then, integrate [hanko-elements](/frontend/elements/README.md) – we provide [example applications](frontend/examples/README.md) and [guides](https://docs.hanko.io/guides/frontend) for your favourite frontend framework in the official documentation

If you want to use the Hanko backend API but prefer to build your own UI, you can still make use of the [hanko-frontend-sdk](/frontend/frontend-sdk/README.md). It forms the basis of our web components, and the client it provides handles communication with the [Hanko backend API](https://docs.hanko.io/api-reference/introduction) and saves you the time of rolling your own.

# Community
## Questions, bugs, ideas
If you have any questions or issues, please check this project's [Q&A section in discussions](https://github.com/teamhanko/hanko/discussions/categories/q-a) and the [open issues](https://github.com/teamhanko/hanko/issues). Feel free to comment on existing issues or create a new issue if you encounter any bugs or have a feature request. For yet unanswered questions, feedback, or new ideas, please open a new discussion.

## Discord community & X
We invite you to join our growing [Discord Community](https://www.hanko.io/community) if you want to get the latest updates on passkeys, WebAuthn, and this project or if you just want to chat with us. You can also [follow us on X](https://x.com/hanko_io).

# Licenses
[hanko-elements](frontend/elements) and [hanko-frontend-sdk](frontend/frontend-sdk) are licensed under the [MIT License](frontend/elements/LICENSE). Everything else in this repository, including [hanko backend](backend), is licensed under the [AGPL-3.0](/LICENSE). Non-Copyleft commercial licensing is available on request.
