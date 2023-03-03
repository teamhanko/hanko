# Hanko Remix example

This is a [Remix](https://remix.run) project bootstrapped with [Create Remix
App](https://www.npmjs.com/package/create-remix).

## Starting the app

### Prerequisites

- a running Hanko API (see the instructions on how to run the API [in
  Docker](../backend/README.md#Docker) or [from Source](../backend/README.md#from-source))
- a running express backend (see the [README](../express) for the express backend)

### Set up environment variables

In the `.env` file set up the correct environment variables:

- `REMIX_APP_HANKO_API`: this is the URL of the Hanko API (default:
  `http://localhost:8000`, can be customized using the `server.public.address` option in
  the [configuration file](../../backend/docs/Config.md))
- `REMIX_APP_TODO_API`: this is the URL of the [express](../express) backend (default:
  `http://localhost:8002`)

### Run development server

The entire process is automated with docker. You can find the instructions in the
[examples/README.md](../README.md#remix) file.
