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

## FlowAPI

The SDK offers a TypeScript-based interface for managing authentication and profile flows with Hanko, enabling the
development of custom frontends with the Hanko FlowAPI. It handles state transitions, action execution, input
validation, and event dispatching, while also providing built-in support for auto-stepping and passkey autofill.
This guide explores its core functionality and usage patterns.

### Initializing a New Flow

Start a new authentication or profile flow using the `createState` method on a Hanko instance. Options allow you to control
event dispatching and auto-step behavior.

```typescript
const state = await hanko.createState("login", {
  dispatchAfterStateChangeEvent: true, // Dispatch after-state-change events by default
  excludeAutoSteps: [], // Empty array means all auto-steps are enabled
});
```

#### Parameters

- **flowName**: The name of the flow (e.g., "login", "register" or "profile").
- **options**:
    - **dispatchAfterStateChangeEvent**: Boolean to enable the `onAfterStateChanged` event after state changes when creating a new state (default: `true`).
    - **excludeAutoSteps**: Array of state names or "all" to skip specific or all auto-steps.

### Understanding the State Object

The `state` object represents the current step in the flow. It contains properties and methods to interact with the flow.

#### Structure

- `state.name`: The current state’s name (e.g., "login_init", "login_password", "success").
- `state.flowName`: The name of the flow (e.g., "login").
- `state.error`: An `Error` object if an action or request fails (e.g., invalid input or network error).
- `state.payload`: State-specific data returned by the API.
- `state.actions`: An object mapping action names to `Action` instances.
- `state.csrfToken`: CSRF token for secure requests.
- `state.status`: HTTP status code of the last response.
- `state.invokedAction`: Name of the last action run on this state (if any).

### Action Availability

Actions can be enabled or disabled based on the backend configuration or the user's state and properties. You can check
whether a specific action is enabled by accessing its `enabled` property:

```typescript
if (state.actions.example_action.enabled) {
  await state.actions.example_action.run();
} else {
  console.log("Action is disabled");
}
```

### Accessing Action Inputs

Each action in `state.actions` has an `inputs` property defining expected input fields.

```typescript
console.log(state.actions.continue_with_login_identifier.inputs);
// Example output:
// {
//   username: {
//     required: true,
//     type: "string",
//     minLength: 3,
//     maxLength: 20,
//     description: "User’s login name"
//   }
// }
```

### Running an Action

Actions transition the flow to a new state. Use the `run` method on an action, passing input values and optional configuration.

#### Basic Example with Type Narrowing

```typescript
if (state.name === "login_init") {
  const newState = await state.actions.continue_with_login_identifier.run({
    username: "user1",
  });
  // Triggers `onBeforeStateChanged` and `onAfterStateChanged` events
  // `newState` is the next state in the flow (e.g., "login_password")
}
```

#### Additional Considerations

- **Type Narrowing**: Check `state.name` to ensure the action exists and inputs are valid for that state.
- **Events**: By default, `run` triggers `onBeforeStateChanged` before the action and `onAfterStateChanged` after the new state is loaded.
- **Validation Errors**: If the action fails due to invalid input (e.g., wrong format or length), `newState.error` will be set to "invalid_form_data", and specific errors will be attached to the related input fields (see "Error Handling" below).

### Event Handlers

The SDK dispatches events via the Hanko instance to track state changes.

#### `onBeforeStateChanged`

Fires before an action is executed, useful for showing loading states.

```typescript
hanko.onBeforeStateChanged(({ state }) => {
  console.log("Action loading:", state.invokedAction);
});
```

#### `onAfterStateChanged`

Fires after a new state is loaded, ideal for rendering UI or handling state-specific logic.

```typescript
hanko.onAfterStateChanged(({ state }) => {
  console.log("Action load finished:", state.invokedAction);

  switch (state.name) {
    case "login_init":
      state.webauthnAutofillActivation(); // Special handler for passkey autofill; requires an <input> field on the page with `autocomplete="username webauthn"` (e.g., <input type="text" name="username" autocomplete="username webauthn" />) so the browser can suggest and autofill passkeys when the user interacts with it.
      break;
    case "login_password":
      // Render password input UI
      if (state.error) {
        console.log("Error:", state.error); // e.g., "invalid_form_data"
      }
      break;
    case "error":
      // Handle network errors or 5xx responses
      console.error("Flow error:", state.error);
      break;
  }
});
```

### Controlling the AfterStateChanged Event

You can disable the automatic `onAfterStateChanged` event and dispatch it manually after custom logic.

```typescript
if (state.name === "login_init") {
  const newState = await state.actions.continue_with_login_identifier.run(
    { username: "user1" },
    { dispatchAfterStateChangeEvent: false }, // Disable automatic dispatch
  );
  // Only `onBeforeStateChanged` is triggered here

  await doSomething(); // Your custom async logic
  newState.dispatchAfterStateChangeEvent(); // Manually trigger the event
}
```

### Auto-Steps

Auto-steps automatically advance the flow for certain states, reducing manual intervention.

#### Supported States

- `preflight`
- `login_passkey`
- `onboarding_verify_passkey_attestation`
- `webauthn_credential_verification`
- `thirdparty`
- `success`
- `account_deleted`

#### Disabling Auto-Steps

Prevent auto-steps by specifying states in `excludeAutoSteps`:

```typescript
const state = await hanko.createState("login", {
  excludeAutoSteps: ["success"], // Skip auto-step for "success"
});
```

#### Manual Auto-Step Execution

```typescript
hanko.onAfterStateChanged(({ state }) => {
  if (state.name === "success") {
    console.log("Flow completed");
    await state.autoStep();
  }
});
```

### Error Handling

#### Input Errors

If an action fails due to invalid inputs:

```typescript
if (state.name === "login_password" && state.error === "invalid_form_data") {
  const passwordError = state.actions.password_login.inputs.password.error;
  console.log("Password error:", passwordError);
}
```

#### Network/API Errors

For network issues or `5xx` responses, the `error` state is entered with details in `state.error`.

### Saving and Loading State

Persist the current flow state to `localStorage` using `save()`.

```typescript
// Save the current state
state.save(); // Stores the state to the localStorage

// Later, recover or start a new flow
const recoveredState = await hanko.createState("login");
```

Please note that the `localStorage` entry will be removed automatically when an action is invoked on the saved state.

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

## License

The `hanko-frontend-sdk` project is licensed under the [MIT License](LICENSE).
