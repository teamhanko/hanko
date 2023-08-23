# End 2 End Tests

This directory contains E2E tests for the Hanko project using [Playwright](https://playwright.dev/).

# Contents

- [Prerequisites](#prerequisites)
  - [Required software](#required-software)
  - [Required services](#required-services)
- [Run the tests](#run-the-tests)
  - [Set up services using Docker Compose](#set-up-services-using-docker-compose)
  - [Install project dependencies](#install-project-dependencies)
  - [Execute the tests](#execute-the-tests)

# Prerequisites

## Required software

To run the tests you need to have the following software installed:

- [Node](https://nodejs.org) v. 18.6.0+/ npm
- [Docker](https://www.docker.com/) / Docker Compose


## Required services

Furthermore, you need running instances of:

- the Hanko [backend](../backend)
- a running frontend application (e.g. our [quickstart](../quickstart)) using the web component provided by
  [hanko-elements](../frontend/elements)
- [Mailslurper](https://github.com/mailslurper/mailslurper) as an SMTP server (used to test passcodes through mail
  retrieval via its [API](https://github.com/mailslurper/mailslurper/wiki/API-Guide))

The tests distinguish between password-based and passwordless scenarios. Each of these requires the proper
[backend](../backend) configuration, i.e. it must be configured to run with either passwords enabled or disabled.

# Run the tests

## Set up services using Docker Compose

To get everything up and running, you can use the existing Docker Compose quickstart in
the [`deploy/docker-compose`](../deploy/docker-compose) directory. From the root project directory, execute:

**Passwords disabled**:

```shell
# compose v1
docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-nopw" up --build

# compose v2
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-nopw" up --build

```

**Passwords enabled**:

```shell
# compose v1
PASSWORD_ENABLED=true docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

# compose v2
PASSWORD_ENABLED=true docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

```

or add the following to the `deploy/docker-compose/config.yaml`

```yaml
password:
    enabled: true
```

and then run

```shell
# compose v1
docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

# compose v2
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

```

## Install project dependencies

Once the services are up and running, install dependencies from inside the `e2e` directory:

`npm install`

`npx playwright install chromium`

## Execute the tests

Then execute the tests using:

**Passwords disabled**:

`npm run test:nopw`

**Passwords enabled**:

`npm run test:pw`

> **Note**: If VSCode is your IDE of choice, you can use
> the [Playwright extension](https://marketplace.visualstudio.com/items?itemName=ms-playwright.playwright) to
> run a test or a group of tests with a single click.

For more information on how to customize npm scripts to run tests using the Playwright CLI please view
the official [documentation](https://playwright.dev/docs/test-cli).


