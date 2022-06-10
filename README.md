[![GoDoc](https://godoc.org/github.com/teamhanko/hanko?status.svg)](https://godoc.org/github.com/teamhanko/hanko)
![Build Status](https://github.com/teamhanko/hanko/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/teamhanko/hanko)](https://goreportcard.com/report/github.com/teamhanko/hanko)

🚧 **PRE-ALPHA** 🚧: This project is not yet functional. We expect to deliver a first usable version in June 2022. If you want to get notified as soon as we're ready you can leave a star, join our [Slack community](https://www.hanko.io/community), or sign up to our [newsletter](https://www.hanko.io/updates).

![Hanko - Open-source user authentication that can do more than just passwords](https://user-images.githubusercontent.com/20115649/167916572-a4d92eaa-7246-4e18-a85d-fc80b4f25c28.svg)

# About Hanko
Hanko is open-source user authentication with a focus on moving the login beyond passwords, while being 100% deployable today – without compromise.

- Passwords, passcodes, and [passkeys](https://www.passkeys.io) API
- JS frontend lib with a feature-complete and highly customizable Hanko web component ("login box")
- JWT issuing
- User management API
- Audit logs API
- FIDO2-certified

[![FIDO2 Certified](https://user-images.githubusercontent.com/20115649/159896561-a94022ba-0e95-417e-807d-b4b7ce19371c.svg)](https://fidoalliance.org/company/hanko/)

Hanko is built and maintained by [Hanko.io](https://www.hanko.io), a startup based in northern Germany, and an active member in the [FIDO Alliance](https://fidoalliance.org/company/hanko/). This project is the sum of 5+ years of experience implementing FIDO and WebAuthn-based authentication in many different applications and platforms.

# We take you on the journey beyond passwords ...
... and make sure your users won't get lost on the way. Passwordless logins have been promised to us for quite some time. But until now, "passwordless" was mostly a compromise that only worked for some of the users and had some severe drawbacks that ultimately led to passwords still being present at almost every login. It's only very recently that the ecosystem of devices, browsers, and operating systems is finally ready to truly move beyond passwords.

With most devices now shipping with biometric sensors like Touch ID, Face ID, and Windows Hello and supporting WebAuthn APIs (web and native) out of the box, it's enabling a truly fascinating login experience that has the potential to replace passwords for good.

# Build your product, not another (password) login
Implementing onboarding and authentication that benefit from end-to-end passwordless and biometric convenience through WebAuthn, but also handle all edge cases and recovery flows is not a simple task.

**That's where Hanko comes in:**

With Hanko, your users will be guided to login to your apps with biometrics instead of passwords. On devices that do not support WebAuthn, or for the first-time login on a new device where no WebAuthn credential is available, passwords or email passcodes can be used. But directly after, the user is always guided to enroll their biometrics.

- A polished, passwordless user experience that does not leave today's users behind
- Biometrics, WebAuthn (passkeys, FIDO Security Keys), passcodes
- Optional password authentication instead of email passcodes to make sure your users won't feel lost
- All edge cases are handled that normally will keep you busy much longer than you would like (e.g., account recovery, unsupported devices, multi-language UI & emails)
- Mobile app support is planned
- Self-hosted or as a lean SaaS hosted by us (coming soon)
- API-first, small footprint, cloud-native

**Build your future-proof web app login in just 5 minutes – with two lines of code – and never look back.**

## Community
Join our [Slack community](https://www.hanko.io/community) if you have any questions about Hanko or just want to chat about authentication, identity, or life in general.

# Quickstart
To try out hanko you can use either docker-compose or kubernetes manifests. For either way you have to first clone this repository with:
```
git clone https://github.com/teamhanko/hanko.git
```

## With docker-compose
Just run:
```
docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --force-recreate
```

After the services are up and running, the example can be opened at `localhost:8888`. To receive emails, without your own
smtp server, we added [mailslurper](https://github.com/mailslurper/mailslurper) which will be available at `localhost:8080`.

> **Note:** Some services are not published to a registry yet and will be built locally before the services are started.

> **Note:** Currently the services are not waiting for postgres to be ready. So the services output an error, that they
> can not connect to the db. Just wait a second until the services are settled.

## With kubernetes
```
kubectl apply -k deploy/k8s/overlays/quickstart
```
TODO describe setup.
# Develop
TODO
