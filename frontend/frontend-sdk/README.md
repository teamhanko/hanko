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
    dispatchAfterStateChangeEvent: true,
    excludeAutoSteps: [],
    loadFromCache: true,
    cacheKey: "hanko-flow-state",
});
```

#### Parameters

- **flowName**: The name of the flow (e.g., "login", "register" or "profile").
- **options**:
    - **dispatchAfterStateChangeEvent**: `boolean` - Whether to dispatch the onAfterStateChange event after state changes (default: true).
    - **excludeAutoSteps**: `AutoStepExclusion` - Array of state names or "all" to skip specific or all auto-steps (default: null).
    - **loadFromCache**: `boolean` - Whether to attempt loading a cached state from localStorage (default: true).
    - **cacheKey**: `string` - The key used for localStorage caching (default: "hanko-flow-state").


### Understanding the State Object

The `state` object represents the current step in the flow. It contains properties and methods to interact with the flow.

#### Structure

- **name**: `StateName` - The current state’s name (e.g., "login_init", "login_password", "success").
- **flowName**: `FlowName` - The name of the flow (e.g., "login").
- **error**: `Error | undefined` - An error object if an action or request fails (e.g., invalid input, network error).
- **payload**: `Payloads[StateName] | undefined` - State-specific data returned by the API.
- **actions**: `ActionMap<StateName>` - An object mapping action names to Action instances.
- **csrfToken**: `string` - CSRF token for secure requests.
- **status**: `number` - HTTP status code of the last response.
- **invokedAction**: `ActionInfo | undefined` - Details of the last action run on this state, if any.
- **previousAction**: `ActionInfo | undefined` - Details of the action that led to this state, if any.
- **isCached**: `boolean` - Whether the state was loaded from localStorage.
- **cacheKey**: `string` - The key used for localStorage caching.
- **excludeAutoSteps**: `AutoStepExclusion` - An array of `StateNames` excluded from auto-stepping.


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
  // Triggers `onBeforeStateChange` and `onAfterStateChange` events
  // `newState` is the next state in the flow (e.g., "login_password")
}
```

#### Additional Considerations

- **Type Narrowing**: Check `state.name` to ensure the action exists and inputs are valid for that state.
- **Events**: By default, `run` triggers `onBeforeStateChange` before the action and `onAfterStateChange` after the new state is loaded.
- **Validation Errors**: If the action fails due to invalid input (e.g., wrong format or length), `newState.error` will be set to "invalid_form_data", and specific errors will be attached to the related input fields (see "Error Handling" below).

### Event Handlers

The SDK dispatches events via the Hanko instance to track state changes.

#### `onBeforeStateChange`

Fires before an action is executed, useful for showing loading states.

```typescript
hanko.onBeforeStateChange(({ state }) => {
  console.log("Action loading:", state.invokedAction);
});
```

#### `onAfterStateChange`

Fires after a new state is loaded, ideal for rendering UI or handling state-specific logic.

```typescript
hanko.onAfterStateChange(({ state }) => {
  console.log("Action load finished:", state.invokedAction);

  switch (state.name) {
    case "login_init":
      state.passkeyAutofillActivation(); // Special handler for passkey autofill; requires an <input> field on the page with `autocomplete="username webauthn"` (e.g., <input type="text" name="username" autocomplete="username webauthn" />) so the browser can suggest and autofill passkeys when the user interacts with it.
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

You can disable the automatic `onAfterStateChange` event and dispatch it manually after custom logic.

```typescript
if (state.name === "login_init") {
  const newState = await state.actions.continue_with_login_identifier.run(
    { username: "user1" },
    { dispatchAfterStateChangeEvent: false }, // Disable automatic dispatch
  );
  // Only `onBeforeStateChange` is triggered here

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
hanko.onAfterStateChange(({ state }) => {
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

```typescript
if (state.name === "error") {
    console.error("Flow error:", state.error);
}
```

### Caching Flow State

Persist and recover flow state using `localStorage`.

#### Saving State

Save the current flow state to `localStorage` using `saveToLocalStorage()`.

```typescript
state.saveToLocalStorage(); // Stores under `state.cacheKey` (default: "hanko-flow-state")
```

Please note that the `localStorage` entry will be removed automatically when an action is invoked on the saved state.

#### Loading State

Recover a cached state or start a new flow:

```typescript
const state = await hanko.createState("login", {
    loadFromCache: true, // Attempts to load from `cacheKey`
    cacheKey: "hanko-flow-state",
});
```

#### Clearing State

Remove the cached state:

```typescript
state.removeFromLocalStorage(); // Deletes from `state.cacheKey`
```

#### Advanced Serialization

For custom persistence:

```typescript
import { State } from "@teamhanko/hanko-frontend-sdk";

const serialized = state.serialize(); // Returns a `SerializedState` object
// Store `serialized` in your storage system

// Later, deserialize it
const recoveredState = await State.deserialize(hanko, serialized, {
    cacheKey: "custom-key",
});
```

This allows integration with other storage mechanisms.

### Session Events

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

### Session Management

The SDK provides methods to manage user sessions and retrieve user information.

#### Getting the User Object

Fetches the current user's profile information.

```typescript
try {
    const user = await hanko.getUser();
    console.log("User profile:", user);
} catch (error) {
    console.error("Failed to fetch user:", error);
    // Handle UnauthorizedError or TechnicalError
}
```

#### Validating a Session

Checks the validity of the current session.

```typescript
try {
    const sessionStatus = await hanko.validateSession();
    console.log("Session status:", sessionStatus);
} catch (error) {
    console.error("Failed to fetch session status:", error);
    // Handle TechnicalError
}
```

#### Getting the Session Token

Retrieves the current session token from the authentication cookie.

```typescript
const token = hanko.getSessionToken();
console.log("Session token:", token);
```

#### Logging out a User

Logs out the current user by invalidating the session.

```typescript
try {
    await hanko.logout();
    console.log("User logged out");
} catch (error) {
    console.error("Failed to fetch user logout:", error);
    // Handle TechnicalError
}
```

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

## License

The `hanko-frontend-sdk` project is licensed under the [MIT License](LICENSE).
