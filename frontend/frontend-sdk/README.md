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

### Options

You can pass certain options, when creating a new `Hanko` instance:

```js
const defaultOptions = {
  timeout: 13000,                                // The timeout (in ms) for the HTTP requests.
  cookieName: "hanko",                           // The cookie name under which the session token is set.
  localStorageKey: "hanko",                      // The prefix / name of the localStorage keys.
  sessionCheckInterval: 30000,                   // Interval (in ms) for session validity checks. Must be greater than 3000 (3s).
  sessionCheckChannelName: "hanko-session-check" // The broadcast channel name for inter-tab communication

};
const hanko = new Hanko("http://localhost:3000", defaultOptions);
```

## Documentation

To see the latest documentation, please click [here](https://teamhanko.github.io/hanko/jsdoc/hanko-frontend-sdk/).

## Exports

### SDK

- `Hanko` - A class that bundles all functionalities.

### Client Classes

- `UserClient` - A class to manage users.
- `ThirdPartyClient` - A class to handle social logins.
- `TokenClient` - A class that handles the exchange of one time tokens for session JWTs.

### Utility Classes

- `WebauthnSupport` - A class to check the browser's WebAuthn support.

### DTO Interfaces

- `PasswordConfig`
- `EmailConfig`
- `AccountConfig`
- `Config`
- `WebauthnFinalized`
- `TokenFinalized`
- `UserInfo`
- `Me`
- `Credential`
- `User`
- `UserCreated`
- `Passcode`
- `WebauthnTransports`
- `Attestation`
- `Email`
- `Emails`
- `WebauthnCredential`
- `WebauthnCredentials`
- `Identity`

### Event Interfaces

- `SessionDetail`

### Event Types

- `CustomEventWithDetail`
- `sessionCreatedType`
- `sessionExpiredType`
- `userLoggedOutType`
- `userDeletedType`

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

### Custom Events

You can bind callback functions to different custom events. The callback function will be called when the event happens
and an object will be passed in, containing event details. The event binding works as follows:

```typescript
// Controls the optional `once` parameter. When set to `true` the callback function will be called only once.
const once = false;

const removeEventListener = hanko.onSessionCreated((eventDetail) => {
    // Your code...
}, once);
```

The following events are available:

- "hanko-session-created": Will be triggered after a session has been created and the user has completed possible
  additional steps (e.g. passkey registration or password recovery). It will also be triggered when the user logs in via
  another browser window. The event can be used to obtain the JWT claims.

```js
hanko.onSessionCreated((sessionDetail) => {
  // A new JWT has been issued.
    console.info("Session created", sessionDetail.claims);
})
```

- "hanko-session-expired": Will be triggered when the session has expired, or when the session has been removed in
  another browser window, because the user has logged out, or deleted the account.

```js
hanko.onSessionExpired(() => {
  // You can redirect the user to a login page or show the `<hanko-auth>` element, or to prompt the user to log in again.
  console.info("Session expired");
})
```

- "hanko-user-logged-out": Will be triggered, when the user actively logs out. In other browser windows, a "hanko-session-expired" event
  will be triggered at the same time.

```js
hanko.onUserLoggedOut(() => {
  // You can redirect the user to a login page or show the `<hanko-auth>` element.
  console.info("User logged out");
})
```

- "hanko-user-deleted": Will be triggered when the user has deleted the account. In other browser windows, a "hanko-session-expired" event
  will be triggered at the same time.

```js
hanko.onUserDeleted(() => {
  // You can redirect the user to a login page or show the `<hanko-auth>` element.
  console.info("User has been deleted");
})
```

Please Take a look into the [docs](https://teamhanko.github.io/hanko/jsdoc/hanko-frontend-sdk/) for more details.

### Translation of outgoing emails

If you use the main `Hanko` client provided by the Frontend SDK, you can use the `lang` parameter in the options when
instantiating the client to configure the language that is used to convey to the Hanko API the
language to use for outgoing emails. If you have disabled email delivery through Hanko and configured a webhook for the
`email.send` event, the value for the `lang` parameter is reflected in the JWT payload of the token contained in the
webhook request in the "Language" claim.

## Bugs

Found a bug? Please report on our [GitHub](https://github.com/teamhanko/hanko/issues) page.

## License

The `hanko-frontend-sdk` project is licensed under the [MIT License](LICENSE).
