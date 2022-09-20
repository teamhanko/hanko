# Hanko Examples

This directory contains examples that show

- integration of web component(s) provided through the`@teamhanko/hanko-elements` package (see [elements](../elements)).
- how to validate JSON Web Tokens (JWT) issued by the Hanko [API](../backend) in a custom backend

It contains:

- an example [express](express) backend - this is a simple version of the well-known todo app
- example frontend applications using the following frameworks:
  - [Angular](angular)
  - [Next.js](nextjs)
  - [React](react)
  - [Vue](vue)

## How to run

1. Start the Hanko API (see the instructions on how to run the API [in Docker](../backend/README.md#Docker) or [from Source](../backend/README.md#from-source))
2. Start the express backend (see the [README](express) for the express backend)
3. Start one of the frontend applications (see the README for the app of your choice)
