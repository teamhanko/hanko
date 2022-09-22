# Hanko Express Example

This directory contains an [express](https://expressjs.com) application that serves as the backend for the example
frontend applications contained in the [examples](../../examples) directory. It is a simple API for creating, listing and
updating "todos".

## Starting the app

### Set up environment variables

In the `.env` file set up the correct environment variables:

- `HANKO_API_URL`: this is the URL of the Hanko API (default: `http://localhost:8000`, can be customized using the `server.public.address` option in the [configuration file](../../backend/docs/Config.md))

### Run the server

Run `npm install` to install dependencies, then run `npm run start` for a development server.
Navigate to `http://localhost:8888/`. The application will automatically reload if you change any of the source files.
