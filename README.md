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
Hanko is an open-source user authentication system with a focus on moving the login beyond passwords, while being 100% deployable today – without compromise.

- Built around latest [passkey](https://www.passkeys.io) technology as introduced by Apple, Google, and Microsoft
- An API for passkeys, email passcodes, and optional password support
- Hanko web component ("login box") that integrates with just 2 lines of code
- API-first, small footprint, cloud-native
- FIDO2-certified

Hanko is built and maintained by [Hanko.io](https://www.hanko.io), an active member of the [FIDO Alliance](https://fidoalliance.org/company/hanko/). This project is the sum of 5 years of experience implementing FIDO and WebAuthn-based authentication in many different applications and platforms.

https://user-images.githubusercontent.com/20115649/194661461-8819db77-4db5-4b24-9859-5a8e68be77fe.mp4

# We take you on the journey beyond passwords ...
... and make sure your users won't get lost on the way. Passwordless logins have been promised to us for quite some time. But until now, "passwordless" was mostly a compromise that only worked for some of the users and had some severe drawbacks that ultimately led to passwords still being present at almost every login. It's only very recently that passkeys were announced and the ecosystem of devices, browsers, and operating systems is finally ready to truly move beyond passwords.

With most devices now shipping with passkey support and biometric sensors like Touch ID, Face ID, and Windows Hello, a truly fascinating login experience is enabled that has the potential to replace passwords for good. Hanko is built for that shift.

**Build your passkey login in just 5 minutes – with two lines of code – and never look back.**

# Roadmap
We are currently in **Beta** and may still have critical bugs. Watch our releases, leave a star, join our [Slack community](https://www.hanko.io/community), or sign up to our [product news](https://www.hanko.io/updates) to follow the development. Here's a brief overview of the current roadmap:

| Status | Feature |
| :---: | :--- |
| ✅ | Passkeys |
| ✅ | Email passcodes |
| ✅ | Passwords |
| ✅ | JWT signing |
| ✅ | User management API |
| ✅ | 📢 Hanko Alpha Release |
| ✅ | `hanko-auth` web component |
| ✅ | Customizable CSS |
| ✅ | 📢 Hanko Beta Release |
| ✅ | JavaScript frontend SDK |
| ⚙️ | Passkey autofill ([Conditional UI](https://github.com/w3c/webauthn/wiki/Explainer:-WebAuthn-Conditional-UI)) |
| ⚙️ | Audit logs API |
| | Rate limiting |
| | `hanko-profile` web component |
| | SSO plugin system (Sign in with Google/Apple/GitHub/...) |
| | Support for native Android and iOS apps |
| | 2FA to secure fallback auth methods |

Additional features that have been requested or that we would like to build but are not yet on the roadmap:
- SMS passcode delivery
- Session management

# Quickstart
The fastest way to try out Hanko is with [docker-compose](https://www.docker.com/products/docker-desktop/).

First you need to clone this repository:
```
git clone https://github.com/teamhanko/hanko.git
```

Then, in the newly created `hanko` folder, just run:
```
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build
```
> **Note**: Docker (Desktop) needs to be running in order for the command to run.

After the services are up and running, the example login can be opened at `localhost:8888`. To receive emails without your own
smtp server, we added [mailslurper](https://github.com/mailslurper/mailslurper) which will be available at `localhost:8080`.

> **Note**: Hanko services are not published to a registry yet and will be built locally before the services are started.

# Monorepo
The Hanko project consists of
- [backend](/backend/README.md) - An authentication API powering passkeys, passcodes, and passwords, as well as user management and JWT token issuing
- [hanko-elements](/elements/README.md) - A slick web component made for Hanko backend that features a polished onboarding and login experience and is highly customizable
- [hanko-frontend-sdk](/frontend-sdk/README.md) - A client package for using the Hanko API
- [example](/example) - The quickstart example app, showing off Hanko's login experience and acting as a reference implementation
- [examples](/examples) - Hanko example implementations for a number of frameworks
- [docs](/docs) - The Hanko documentation

# Community
Join our [Slack community](https://www.hanko.io/community) if you have any questions about Hanko or just want to chat about passkeys, authentication, identity, or life in general. You can also [follow us on Twitter](https://twitter.com/hanko_io) or just [reach out via email](https://www.hanko.io/contact).

# Licenses
[hanko-elements](elements) and [hanko-frontend-sdk](frontend-sdk) are licensed under the [MIT License](elements/LICENSE). Everything else in this repository, including [hanko backend](backend), is licensed under the [AGPL-3.0](/LICENSE).
