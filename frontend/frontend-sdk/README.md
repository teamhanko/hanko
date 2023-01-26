# hanko-frontend-sdk

This package utilizes the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md) to provide
functionality that allows an easier UI integration. It is meant for use in browsers only.

## Installation

```shell
# npm
npm install @teamhanko/hanko-frontend-sdk

# yarn
yarn add @teamhanko/hanko-frontend-sdk

# pnpm
pnpm install @teamhanko/hanko-frontend-sdk
```

## Usage

Import as a module:

```typescript
import { Hanko } from "@teamhanko/hanko-frontend-sdk"

const hanko = new Hanko("http://localhost:3000")
```

With a script tag via CDN:

```html
<script src="https://cdn.jsdelivr.net/npm/@teamhanko/hanko-frontend-sdk/dist/sdk.umd.js"></script>

<script>
    const hanko = new hankoFrontendSdk.Hanko("http://localhost:3000")
    ...
</script>
```

## Documentation

To see the latest documentation, please click [here](https://docs.hanko.io/jsdoc/hanko-frontend-sdk).

## Exports

### SDK

- `Hanko` - A class that bundles all functionalities.

### Clients

- `ConfigClient` - A class to fetch configurations.
- `UserClient` - A class to manage users.
- `WebauthnClient` - A class to handle WebAuthn-related functionalities.
- `PasswordClient` - A class to manage passwords and password logins.
- `PasscodeClient` - A class to handle passcode logins.

### Utilities

- `WebauthnSupport` - A class to check the browser's WebAuthn support.

### DTOs

- `Config`
- `PasswordConfig`
- `WebauthnFinalized`
- `Credential`
- `UserInfo`
- `User`
- `Email`
- `Emails`
- `Passcode`

### Errors

- `HankoError`
- `TechnicalError`
- `ConflictError`
- `RequestTimeoutError`
- `WebauthnRequestCancelledError`
- `InvalidPasswordError`
- `InvalidPasscodeError`
- `InvalidWebauthnCredentialError`
- `PasscodeExpiredError`
- `MaxNumOfPasscodeAttemptsReachedError`
- `NotFoundError`
- `TooManyRequestsError`
- `UnauthorizedError`

## Examples

### Get the current user / Validate the JWT against the Hanko API

The Hanko API issues a JWT when a user logs in. For certain actions, like obtaining the user object, a valid  JWT is
required. The following example shows how to get the user object of the current user, or to identify that the user is
not logged in:

```typescript
import { Hanko, UnauthorizedError } from "@teamhanko/hanko-frontend-sdk"

const hanko = new Hanko("http://localhost:3000")

try {
    const user = await hanko.user.getCurrent()

    // A valid JWT is in place so that the user object was able to be fetched.
} catch (e) {
    if (e instanceof UnauthorizedError) {
        // Display an error or prompt the user to login again. After a successful call to `hanko.webauthn.login()`,
        // `hanko.password.login()` or `hanko.passcode.finalize()` a JWT will be issued and `hanko.user.getCurrent()`
        // would succeed.
    }
}
```

### Register a WebAuthn credential

There are a number of situations where you may want the user to register a WebAuthn credential. For example, after user
creation, when a user logs in to a new browser/device, or to take advantage of the "caBLE" support and pair a smartphone
with a desktop computer:

```typescript
import { Hanko, UnauthorizedError, WebauthnRequestCancelledError } from "@teamhanko/hanko-frontend-sdk"

const hanko = new Hanko("http://localhost:3000")

// By passing the user object (see example above) to `hanko.webauthn.shouldRegister(user)` you get an indication of
// whether a WebAuthn credential registration should be performed on the current browser. This is useful if the user has
// logged in using a method other than WebAuthn, and you then want to display a UI that allows the user to register a
// credential when possibly none exists.

try {
    // Will cause the browser to present a dialog with various options depending on the WebAuthn implemention.
    await hanko.webauthn.register()

    // Credential has been registered.
} catch(e) {
    if (e instanceof WebauthnRequestCancelledError) {
        // The WebAuthn API failed. Usually in this case the user cancelled the WebAuthn dialog.
    } else if (e instanceof UnauthorizedError) {
        // The user needs to login to perform this action.
    }
}
```

## Bugs

Found a bug? Please report on our [GitHub](https://github.com/teamhanko/hanko/issues) page.

## License

The `hanko-frontend-sdk` project is licensed under the [MIT License](LICENSE).
