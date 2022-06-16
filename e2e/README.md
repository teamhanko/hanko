# End 2 End Tests

This directory contains E2E tests for the Hanko project using [Playwright](https://www.https://playwright.dev/).

## Running

Running the tests requires:

- [Node](https://nodejs.org) / npm
- [Docker](https://www.docker.com/) / Docker Compose
- a running Hanko [backend](../backend)
- a running frontend application (e.g. our [example](../example)) using the web component provided by
  [hanko-js](../hanko-js)
- [Mailslurper](https://github.com/mailslurper/mailslurper) as an SMTP server (used to test Passcodes through mail
  retrieval via its [API](https://github.com/mailslurper/mailslurper/wiki/API-Guide))

The tests distinguish between password-based and passwordless scenarios. Each of these requires the proper
[backend](../backend) configuration, i.e. it must be configured to run with either passwords enabled or disabled. To get
everything up and running, you can use the existing Docker Compose quickstart in
the [`deploy/docker-compose`](..deploy/docker-compose) directory. From the root project directory, execute:

**Passwords disabled**:

```shell
# compose v1
docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-nopw" up --build

# compose v2
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-nopw" up --build

```

**Passwords enabled**:

`PASSWORD_ENABLED=true docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build`

```shell
# compose v1
PASSWORD_ENABLED=true docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

# compose v2
PASSWORD_ENABLED=true docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

```

or add the following to the `deploy/docker-compose/config.yaml`

```yaml
password:
    enabled: false
```

and then run

```shell
# compose v1
docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-nopw" up --build

# compose v2
docker compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart-pw" up --build

```

Once the services are up and running, execute the tests from the `e2e` directory:

**Passwords disabled**:

`npm run test:nopw`

**Passwords enabled**:

`npm run test:pw`

> **Note**: If VSCode is your IDE of choice, you can use
> the [Playwright extension](https://marketplace.visualstudio.com/items?itemName=ms-playwright.playwright) to
> run a test or a group of tests with a single click.

For more information on how to customize npm scripts to run tests using the Playwright CLI please view
the official [documentation](https://playwright.dev/docs/test-cli).


