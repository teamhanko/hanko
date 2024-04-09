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
Hanko is an open-source authentication and user management solution with a focus on moving the login beyond passwords while being 100% deployable today.

- Built for [passkeys](https://www.passkeys.io) as introduced by Apple, Google, and Microsoft
- Fast integration with Hanko Elements web components (login box and user profile)
- API-first, small footprint, cloud-native

Available for self-hosting and on [Hanko Cloud](https://www.hanko.io).

Hanko is built and maintained by [Hanko.io](https://www.hanko.io), an active member of the [FIDO Alliance](https://fidoalliance.org/company/hanko/).

# We take you on the journey beyond passwords ...
... and make sure your users won't get lost on the way. Passwordless logins have been promised to us for quite some time. But until now, "passwordless" was mostly a compromise that only worked for some of the users and had some severe drawbacks that ultimately led to passwords still being present at almost every login. It's only very recently that passkeys were announced, and the ecosystem of devices, browsers, and operating systems is finally ready to truly move beyond passwords.

With most devices and browsers now shipping with passkey support and convenient built-in authentication technology like Touch ID, Face ID, and Windows Hello, a much better login experience is enabled that will replace passwords for good. Hanko is built for that shift.

**Build your passkey-powered auth stack with a few lines of code – and never look back.**

# Architecture
The main building blocks of the Hanko project are
- [backend](/backend/README.md) - An authentication API for passkeys, passcodes, and (optional) passwords, OAuth SSO, as well as user management and JWT issuing
- [hanko-elements](/frontend/elements/README.md) - Web components made for Hanko backend that provide onboarding and login functionality and are customizable with CSS
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
4. if you have an [enterprise license](/backend/ee) or use [Hanko Cloud](https://cloud.hanko.io) you can also integrate [SAML SSO](/docs). Feel free to use [this guide](https://docs.hanko.io/guides/frontend) to start with SAML SSO

If you want to use the Hanko backend API but prefer to build your own UI, you can still make use of the [hanko-frontend-sdk](/frontend/frontend-sdk/README.md). It forms the basis of our web components, and the client it provides handles communication with the [Hanko backend API](https://docs.hanko.io/api-reference/introduction) and saves you the time of rolling your own.

# Contact us
Schedule a Hanko demo. Learn how Hanko will speed up your registration and login flows with passkeys.

<a target="_blank" href="https://cal.com/team/hanko/demo"><img alt="Book us with Cal.com"  src="https://cal.com/book-with-cal-light.svg" /></a>

# Roadmap
Watch our releases, leave a star, join our [Discord community](https://www.hanko.io/community), or sign up to our [product news](https://www.hanko.io/updates) to follow the development. Here's a brief overview of our current roadmap:

| Status | Feature |
|:------:| :--- |
|✅| Passkeys |
|✅| Email passcodes |
|✅| Passwords |
|✅| JWT signing |
|✅| User management API |
|✅| 📢 Hanko Alpha Release |
|✅| `<hanko-auth>` web component |
|✅| Customizable CSS |
|✅| 📢 Hanko Beta Release |
|✅| JavaScript frontend SDK |
|✅| Passkey autofill ([Conditional UI](https://github.com/w3c/webauthn/wiki/Explainer:-WebAuthn-Conditional-UI)) |
|✅| Audit logs API |
|✅| Security Key support |
|✅| Mobile app support |
|✅| `<hanko-profile>` web component |
|✅| Rate limiting |
|✅| OAuth logins (Sign in with Apple/Google/GitHub) |
|✅| i18n & custom translations |
|✅| User import |
|✅| Disable sign-ups |
|✅| User export |
|✅| SAML Enterprise SSO |
|✅| Webhooks |
|⚙️| API-supported auth flows |
|⚙️| Passkey-only and OAuth-only configurations |
|⚙️| Username support (non-email) |
|⚙️| Optional / user-deleteable passwords |
|⚙️| OIDC Enterprise SSO |
|⚙️| 2FA (TOTP, security keys) |
| | Email templates & i18n |
| | Refresh tokens / sessions |
| | 📢 Hanko 1.0 Release |
| | `<hanko-menu>` web component |
| | Custom Social SSO connections (OIDC/OAuth2) |
| | Email security notifications |
| | Custom JWT claims |
| | Mobile SDKs |
| | SMS passcodes |

Additional features that have been requested or that we would like to build but are currently not on the roadmap:
- Custom user data / fields
- Privileged sessions & step-up authentication
- Bot protection / CAPTCHA
- Hosted auth pages / OIDC provider

# Community
## Questions, bugs, ideas
If you have any questions or issues, please check this project's [Q&A section in discussions](https://github.com/teamhanko/hanko/discussions/categories/q-a) and the [open issues](https://github.com/teamhanko/hanko/issues). Feel free to comment on existing issues or create a new issue if you encounter any bugs or have a feature request. For yet unanswered questions, feedback, or new ideas, please open a new discussion.

## Discord community & X
We invite you to join our growing [Discord community](https://www.hanko.io/community) if you want to get the latest updates on passkeys, WebAuthn, and this project or if you just want to chat with us. You can also [follow us on X](https://x.com/hanko_io).

# Licenses
[hanko-elements](frontend/elements) and [hanko-frontend-sdk](frontend/frontend-sdk) are licensed under the [MIT License](frontend/elements/LICENSE). Everything else in this repository, including [hanko backend](backend), is licensed under the [AGPL-3.0](/LICENSE).
