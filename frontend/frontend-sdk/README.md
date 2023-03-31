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

### Client Classes

- `ConfigClient` - A class to fetch configurations.
- `UserClient` - A class to manage users.
- `WebauthnClient` - A class to handle WebAuthn-related functionalities.
- `PasswordClient` - A class to manage passwords and password logins.
- `PasscodeClient` - A class to handle passcode logins.
- `ThirdPartyClient` - A class to handle social logins.
- `TokenClient` - A class that handles the exchange of one time tokens for session JWTs.

### Utility Classes

- `WebauthnSupport` - A class to check the browser's WebAuthn support.

### DTO Interfaces

- `Config`
- `PasswordConfig`
- `WebauthnFinalized`
- `Credential`
- `UserInfo`
- `User`
- `Email`
- `Emails`
- `Passcode`

### Event Detail Interfaces

- `SessionCreatedEventDetail`
- `AuthFlowCompletedEventDetail`

### Error Classes

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

const hanko = new Hanko("https://[HANKO_API_URL]")

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

const hanko = new Hanko("https://[HANKO_API_URL]")

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

### Events

To build dynamic websites or handle the JWT by your own, you may want to utilize available events as follows:

```typescript
import { Hanko } from "@teamhanko/hanko-frontend-sdk"

const hanko = new Hanko("https://[HANKO_API_URL]")

hanko.onSessionCreated((detail) => {
    // Executes when there already is a session, after the user signs in, or when the JWT has been updated. It will
    // work across browser windows and you can obtain the JWT from the detail object, if you need to manage it by your
    // own. Please note, that the JWT is only available, when the Hanko API configuration allows to obtain the JWT.
    console.log(`User signed in (userID: ${detail.userID}, jwt: ${detail.jwt})`);
})

hanko.onAuthFlowCompleted((detail) => {
    // Login or registration has been finished through the `<hanko-auth>` element. You can now redirect the user to a
    // secured page or fetch secured content in use of the previously issued JWT.
    console.log(`Authentication flow completed (userID: ${detail.userID})`);
})

hanko.onSessionRemoved(() => {
    // Executes across all browser windows after the session has expired. The user can now be redirected back to a
    // login page.
    console.log("User logged out or session has expired");
})

hanko.onUserDeleted(() => {
    // Executes after the user deleted the account. The user can be redirected to a "goodbye" or back to a login page.
    console.log("User deleted");
})
```

Take a look into the docs for more details.

## Bugs

Found a bug? Please report on our [GitHub](https://github.com/teamhanko/hanko/issues) page.

## License

The `hanko-frontend-sdk` project is licensed under the [MIT License](LICENSE).
