# hanko-frontend-sdk

This package utilizes the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md) to provide basic
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
```

With a script tag via CDN:

```html
<script type="module" src="https://unpkg.com/@teamhanko/hanko-client/dist/sdk.js">
```

## Documentation

To see documentation, please click [here](https://teamhanko.github.io/hanko-frontend-sdk).

## Exports

### SDK

- `Hanko` - A class that bundles all functionalities.

### Clients

- `ConfigClient` - A class to fetch configurations.
- `UserClient` - A class to manage users.
- `WebauthnClient` - A class to handle WebAuthN-related functionalities.
- `PasswordClient` - A class to manage passwords and password logins.
- `PasscodeClient` - A class to handle passcode logins.

### Utilities

- `WebauthnSupport` - A class to check the browser's WebAuthN support.

### DTOs

- `Config`
- `PasswordConfig`
- `WebauthnFinalized`
- `Credential`
- `UserInfo`
- `User`
- `Passcode`

### Errors

- `HankoError`
- `TechnicalError`
- `ConflictError`
- `RequestTimeoutError`
- `WebAuthnRequestCancelledError`
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

The Hanko API issues a JWT when a user logs in. For certain actions, like obtaining the user object, a valid
JWT is required. The SDK will handle authorization cookies or headers in the background:

```typescript
import { Hanko, UnauthorizedError } from "@teamhanko/hanko-frontend-sdk"

const hanko = new Hanko("http://localhost:3000")

try {
    const user = await hanko.user.getCurrent()

    // A valid JWT is in place so that the user object was able to be fetched.
    console.info(user.id, user.email)
} catch (e) {
    if (e instanceof UnauthorizedError) {
        // The user needs to login (use for example: `hanko.webauthn.login()`) to perform this action.
    }
}
```

### Register a WebAuthN credential

```typescript
import { Hanko, UnauthorizedError, WebAuthnRequestCancelledError } from "@teamhanko/hanko-frontend-sdk"

const hanko = new Hanko("http://localhost:3000")

try {
    await hanko.webauthn.register()
    // Credential has been registered.
} catch(e) {
    if (e instanceof WebAuthnRequestCancelledError) {
        // The WebAuthN API failed. Usually in this case the user aborted the WebAuthN dialog or there was no
        // suitable credential.
    } else if (e instanceof UnauthorizedError) {
        // The user needs to login again to perform this action.
    }
}
```
