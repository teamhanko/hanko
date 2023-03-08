# Hanko Svelte example

## Starting the app

### Prerequisites

- a running Hanko API (see the instructions on how to run the API [in Docker](../../backend/README.md#Docker) or [from Source](../../backend/README.md#from-source))
- a running express backend (see the [README](../express) for the express backend)

### Set up environment variables

In the `.env` file set up the correct environment variables:

- `VITE_HANKO_API`: this is the URL of the Hanko API (default: `http://localhost:8000`, can be customized using the `server.public.address` option in the [configuration file](../../backend/docs/Config.md))
- `VITE_TODO_API`: this is the URL of the [express](../express) backend (default: `http://localhost:8002`)

### Run development server

Run `npm install` to install dependencies, then run `npm run start` for a development server. Navigate to `http://localhost:8888/`. The application will automatically reload if you change any of the source files.
