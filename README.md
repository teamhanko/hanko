[![GoDoc](https://godoc.org/github.com/teamhanko/hanko?status.svg)](https://godoc.org/github.com/teamhanko/hanko)
![Build Status](https://github.com/teamhanko/hanko/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/teamhanko/hanko)](https://goreportcard.com/report/github.com/teamhanko/hanko)

![Hanko - Open-source user authentication that can do more than just passwords](https://user-images.githubusercontent.com/20115649/167916572-a4d92eaa-7246-4e18-a85d-fc80b4f25c28.svg)

# About Hanko
Hanko is open-source user authentication with a focus on moving the login beyond passwords, while being 100% deployable today – without compromise.

- Fully built around latest [passkey](https://www.passkeys.io) technology as introduced by Apple, Google, and Microsoft
- Also supports passwords and email passcodes
- JS frontend lib with a highly customizable Hanko web component ("login box")
- JWT issuing
- User management API
- Audit logs API (coming soon)
- Mobile SDKs (coming soon)
- Self-hosted or Hanko Cloud (coming soon)
- API-first, small footprint, cloud-native
- FIDO2-certified

[![FIDO2 Certified](https://user-images.githubusercontent.com/20115649/159896561-a94022ba-0e95-417e-807d-b4b7ce19371c.svg)](https://fidoalliance.org/company/hanko/)

Hanko is built and maintained by [Hanko.io](https://www.hanko.io), an active member of the [FIDO Alliance](https://fidoalliance.org/company/hanko/). This project is the sum of more than 5 years of experience implementing FIDO and WebAuthn-based authentication in many different applications and platforms.

# We take you on the journey beyond passwords ...
... and make sure your users won't get lost on the way. Passwordless logins have been promised to us for quite some time. But until now, "passwordless" was mostly a compromise that only worked for some of the users and had some severe drawbacks that ultimately led to passwords still being present at almost every login. It's only very recently that passkeys were announced and the ecosystem of devices, browsers, and operating systems is finally ready to truly move beyond passwords.

With most devices now shipping with passkey support and biometric sensors like Touch ID, Face ID, and Windows Hello, a truly fascinating login experience is enabled that has the potential to replace passwords for good. Hanko is built for that shift.

# Build your product, not another (password) login
Implementing onboarding and authentication that benefit from end-to-end passwordless and biometric convenience with passkeys and WebAuthn, but also handle all edge cases and recovery flows is not a simple task.

**That's where Hanko comes in:**

With Hanko, your users will be guided to login to your apps with passkeys and biometrics instead of passwords. On devices that do not support passkeys, or for the first-time login on a new device where no passkey is available, email passcodes or passwords (optional) can be used. But directly after, the user is always guided to create a passkey.

- A polished, passwordless user experience with optional password authentication to make sure your users won't feel lost
- A slick Hanko web component makes integrating the Hanko login experience into your app possible with just two lines of code
- All edge cases are handled that normally will keep you busy much longer than you would like (e.g., account recovery, unsupported devices, multi-language UI & emails)

**Build your passkey login in just 5 minutes – with two lines of code – and never look back.**

# Roadmap
This project is in **Beta** and may still have critical bugs. Leave a star, join our [Slack community](https://www.hanko.io/community), or sign up to our [newsletter](https://www.hanko.io/updates) to follow the development. Here's a brief overview of our current roadmap:
| Status | Feature |
| :---: | :--- |
| ✅ | Passkeys |
| ✅ | Passcodes |
| ✅ | Passwords |
| ✅ | JWT issuing |
| ✅ | Admin API |
| ✅ | Hanko web component (hanko-js) |
| 🏗️ | Custom CSS |
| 🏗️ | Events / Audit logs API |
| | ***Hanko 1.0 Release*** |
| | 2FA with TOTP and FIDO Security Keys |
| | Mobile app support |
| | Sign in with Google/Apple/GitHub |

# Community
Join our [Slack community](https://www.hanko.io/community) if you have any questions about Hanko or just want to chat about passkeys, authentication, identity, or life in general. You can also [follow us on Twitter](https://twitter.com/hanko_io) or just [reach out via email](https://www.hanko.io/contact).

# Quickstart
To try out hanko you can use docker-compose. First you need to clone this repository with:
```
git clone https://github.com/teamhanko/hanko.git
```

## With docker-compose
Just run:
```
docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --force-recreate
```

After the services are up and running, the example login can be opened at `localhost:8888`. To receive emails without your own
smtp server, we added [mailslurper](https://github.com/mailslurper/mailslurper) which will be available at `localhost:8080`.

> **Note:** Services are not published to a registry yet and will be built locally before the services are started.

> **Note:** Currently the services are not waiting for postgres to be ready. This sometimes results in error outputs by the services, indicating that they
> cannot connect to the database. As soon as postgres is ready, the services reconnect on their own.
