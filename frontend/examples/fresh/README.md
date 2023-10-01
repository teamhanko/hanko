# Hanko Fresh example

This is a [Fresh](fresh.deno.dev/) project.

## Starting the app

### Prerequisites

- a running Hanko API (see the instructions on how to run the API [in Docker](../../../backend/README.md#Docker) or [from Source](../../../backend/README.md#from-source))
- a `Deno` installation

### Set up environment variables

In the `config.ts` file set up the correct variables:

- `HANKO_API_URL`: this is the URL of the Hanko API (default: `http://localhost:8000`)

### Run development server

Run `deno task start` for a development server. Navigate to `http://localhost:8888/`. This will watch the project directory and restart as necessary.
