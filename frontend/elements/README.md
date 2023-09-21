# Hanko Elements

Provides web components that will bring a modern login and registration experience
to your users. It integrates the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md), a backend
that provides the underlying functionalities.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Importing the Module](#importing-the-module)
  - [Registering the Web Components](#registering-the-web-components)
  - [Embedding the Web Components](#embedding-the-web-components)
  - [Using the Frontend-SDK](#using-the-frontend-sdk)
- [UI Customization](#ui-customization)
  - [CSS Variables](#css-variables)
  - [CSS Shadow Parts](#css-shadow-parts)
  - [CSS Classes (not recommended)](#css-classes-not-recommended)
- [Translations](#translations)
  - [Default Behavior](#default-behavior)
  - [Installing Additional Translations](#installing-additional-translations)
  - [Modifying Translations](#modifying-translations)
  - [Adding New Translations](#adding-new-translations)
  - [Using External Files](#using-external-files)
  - [Fallback Language](#fallback-language)
- [Experimental Features](#experimental-features)
  - [Conditional Mediation / Autofill assisted Requests](#conditional-mediation--autofill-assisted-requests)
- [Live Demo](#live-demo)
- [Examples](#examples)
- [Frontend framework integrations](#frontend-framework-integrations)
- [Exports](#exports)
  - [Functions](#functions)
  - [Interfaces](#interfaces)
- [Browser support](#browser-support)
- [Bugs](#bugs)
- [License](#license)

## Features

- User Authentication: Provides a secure and user-friendly way to handle user authentication within web applications.
- Authentication Flows: Supports various authentication flows, including passwordless authentication and biometric
  authentication.
- Web Component Library: Offers customizable web components that can be easily integrated into web applications.
- Profile Management: Allows users to view and manage their profile information through the profile component.
- Event Handling: Provides event listeners for authentication and session-related events, enabling customization and
  control over the user experience.
- Localization and Internationalization: Supports multiple languages and provides translation options for a global user
  base.
- Integration Flexibility: Offers versatile choices for integration, including CDN or npm. It supports both TypeScript
  and non-TypeScript environments, allowing flexibility based on the project requirements.
- Customization: Enables customization of visual styles, branding, and user interface elements to align with the overall
  application design.
- Documentation and Support: Offers documentation, example apps, frontend framework integration guides and support via
  Slack to assist with integration and troubleshooting.

## Installation

To use the Hanko Elements module in your project, you can install it via npm, yarn, or pnpm. Alternatively, you can also
[import the module](#importing-the-module) directly via a CDN.

```shell
# npm
npm install @teamhanko/hanko-elements

# yarn
yarn add @teamhanko/hanko-elements

# pnpm
pnpm install @teamhanko/hanko-elements
```

## Usage

To integrate Hanko, you need to import and call the `register()` function from the `hanko-elements` module. Once this is
done, you can use the web components in your HTML code. For a functioning page, at least the `<hanko-auth>` element
should be placed, so the users can sign in, and also, a handler for the "onAuthFlowCompleted" event should be added, to
customize the behaviour after the authentication flow has been completed (e.g. redirect to another page). These steps
will be described in the following sections.

### Importing the Module

To use the web components, you need to register them using the `register()` function provided by the `hanko-elements`
package.

If you're using a module bundler like webpack or Parcel, you can import the `register()` function from the
`@teamhanko/hanko-elements` package in your TypeScript or JavaScript file:

```typescript
import { register } from "@teamhanko/hanko-elements";
```

If you prefer using a CDN, you can include a script tag with the import statement pointing to the CDN URL where the
`hanko-elements` package is hosted:

```html
<script type="module">
  import { register } from "https://cdn.jsdelivr.net/npm/@teamhanko/hanko-elements/dist/elements.js";
</script>
```

### Registering the Web Components

After importing the `register()` function, call it with the URL of the Hanko API as an argument to register the Hanko
elements with the browser's `CustomElementRegistry`.

```javascript
const { hanko } = await register("https://hanko.yourdomain.com");
```

You can also pass certain options:

```javascript
const defaultOptions = {
  shadow: true,                    // Set to false if you do not want the web component to be attached to the shadow DOM.
  injectStyles: true,              // Set to false if you do not want to inject any default styles.
  enablePasskeys: true,            // Set to false if you do not want to display passkey-related content.
  hidePasskeyButtonOnLogin: false, // Hides the button to sign in with a passkey on the login page.
  translations: null,              // Additional translations can be added here. English is used when the option is not
                                   // present or set to `null`, whereas setting an empty object `{}` prevents the elements
                                   // from displaying any translations.
  translationsLocation: "/i18n",   // The URL or path where the translation files are located.
  fallbackLanguage: "en",          // The fallback language to be used if a translation is not available.
  storageKey: "hanko",             // The name of the cookie the session token is stored in and the prefix / name of local storage keys
};

const { hanko } = await register(
  "https://hanko.yourdomain.com",
  defaultOptions
);
```

Replace "https://hanko.yourdomain.com" with the actual URL of your Hanko API.

### Embedding the Web Components

If you have followed the steps mentioned above, you should now be able to place the web components anywhere in the body
of your HTML. A minimal example would look like this:

```html
<hanko-auth id="authComponent"></hanko-auth>

<script type="module">
  import { register } from "https://cdn.jsdelivr.net/npm/@teamhanko/hanko-elements/dist/elements.js";

  await register("https://hanko.yourdomain.com");

  const authComponent = document.getElementById("authComponent");
  authComponent.addEventListener("onAuthFlowCompleted", () => {
    // redirect to a different page
  });
</script>
```

The individual web component are described in the following sections.

#### &lt;hanko-auth&gt;

A web component that handles user login and user registration.

##### Markup

```html
<hanko-auth></hanko-auth>
```

##### Attributes

- `prefilled-email` Used to prefill the email input field.
- `lang` Used to specify the language of the content within the element. See [Translations](#translations).
- `experimental` A space-separated list of experimental features to be enabled.
  See [experimental features](#experimental-features).

#### &lt;hanko-profile&gt;

A web component that allows to manage emails, passwords and passkeys.

##### Markup

```html
<hanko-profile></hanko-profile>
```

##### Attributes

- `lang` Used to specify the language of the content within the element. See [Translations](#translations).

#### &lt;hanko-events&gt;

A web component that allows to bind event handler to certain events, without displaying UI elements. Events can be
subscribed to with the `<hanko-auth>` and `<hanko-profile>` components in the same manner. Also, you can bind event
handler via the `frontend-sdk` (see next section).

##### Markup

```html
<hanko-events id="events"></hanko-events>
<script>
  document
    .getElementById("events")
    .addEventListener("onAuthFlowCompleted", console.log);
  // more events are available (see "frontend-sdk" docs)...
</script>
```

### Using the Frontend-SDK

The following examples will cover some common use-cases for the `hanko-frontend-sdk` instance returned by
the `register()` function, but please take a look into the
[frontend-sdk docs](https://docs.hanko.io/jsdoc/hanko-frontend-sdk/) for details.

Note that you can create a `hanko-frontend-sdk` instance without having to register the web components as follows:

```js
import { Hanko } from "@teamhanko/hanko-elements";

const hanko = new Hanko("https://hanko.yourdomain.com");
```

#### Events

It is possible to bind callbacks to different custom events in use of the SDKs event listener functions.
The callback function will be called when the event happens and an object will be passed in, containing event details.

##### Auth Flow Completed

Will be triggered after a session has been created and the user has completed possible
additional steps (e.g. passkey registration or password recovery) via the `<hanko-auth>` element.

```js
hanko.onAuthFlowCompleted((authFlowCompletedDetail) => {
  // Login, registration or recovery has been completed successfully. You can now take control and redirect the
  // user to protected pages.
  console.info(
    `User successfully completed the registration or authorization process (user-id: "${authFlowCompletedDetail.userID}")`
  );
});
```

##### Session Created

Will be triggered before the "hanko-auth-flow-completed" happens, as soon as the user is technically logged in. It will
also be triggered when the user logs in via another browser window. The event can be used to obtain the JWT.

Please note, that the JWT is only available, when the Hanko-API configuration allows to obtain the JWT. When using
Hanko-Cloud the JWT is always present, for self-hosted Hanko-APIs you can restrict the cookie to be readable by the
backend only, as long as your backend runs under the same domain as your frontend. To do so, make sure the config
parameter "session.enable_auth_token_header" is turned off via the
[Hanko-API configuration](https://github.com/teamhanko/hanko/blob/main/backend/docs/Config.md). If you want the JWT to
be contained in the event details, you need to turn on "session.enable_auth_token_header" when using a cross-domain
setup. When it's a same-domain setup you need to turn off "session.cookie.http_only" to make the JWT accessible to the
frontend.

```js
hanko.onSessionCreated((sessionDetail) => {
  // A new JWT has been issued.
  console.info(
    `Session created or updated (user-id: "${sessionDetail.userID}", jwt: ${sessionDetail.jwt})`
  );
});
```

##### Session Expired

Will be triggered when the session has expired, or when the session has been removed in another browser window, because
the user has logged out, or deleted the account.

```js
hanko.onSessionExpired(() => {
  // You can redirect the user to a login page or show the `<hanko-auth>` element, or to prompt the user to log in again.
  console.info("Session expired");
});
```

##### User Logged Out

Will be triggered, when the user actively logs out. In other browser windows, a "hanko-session-expired" event will be
triggered at the same time.

```js
hanko.onUserLoggedOut(() => {
  // You can redirect the user to a login page or show the `<hanko-auth>` element.
  console.info("User logged out");
});
```

##### User Deleted

Will be triggered when the user has deleted the account. In other browser windows, a "hanko-session-expired" event will
be triggered at the same time.

```js
hanko.onUserDeleted(() => {
  // You can redirect the user to a login page or show the `<hanko-auth>` element.
  console.info("User has been deleted");
});
```

To learn what else you can do, check out the
[custom-events](https://github.com/teamhanko/hanko/tree/main/frontend/frontend-sdk#custom-events) README.

#### Sessions

Determine whether the user is logged in:

```js
hanko.session.isValid();
```

Getting the session details:

```js
const session = hanko.session.get();

if (session) {
  console.info(`userID: ${session.userID}, jwt: ${session.jwt}`);
}
```

#### User Client

The SDK contains several client classes to make the communication with the Hanko-API easier. Here some examples of
things you might want to do:

Getting the current user:

```js
const user = await hanko.user.getCurrent();
console.info(`id: ${user.id}, email: ${user.email}`);
```

Log out a user:

```js
await hanko.user.logout();
```

To learn how error handling works and what else you can do with SDK, take a look into
the [frontend-sdk docs](https://docs.hanko.io/jsdoc/hanko-frontend-sdk/).

## UI Customization

### CSS Variables

CSS variables can be used to style the `hanko-auth` and `hanko-profile` elements to your needs. A list of all CSS
variables including default values can be found below:

```css
hanko-auth,
hanko-profile {
  /* Color Scheme */
  --color: #171717;
  --color-shade-1: #8f9095;
  --color-shade-2: #e5e6ef;

  --brand-color: #506cf0;
  --brand-color-shade-1: #6b84fb;
  --brand-contrast-color: white;

  --background-color: white;
  --error-color: #e82020;
  --link-color: #506cf0;

  /* Font Styles */
  --font-weight: 400;
  --font-size: 16px;
  --font-family: sans-serif;

  /* Border Styles */
  --border-radius: 8px;
  --border-style: solid;
  --border-width: 1px;

  /* Item Styles */
  --item-height: 42px;
  --item-margin: 0.5rem 0;

  /* Container Styles */
  --container-padding: 30px;
  --container-max-width: 410px;

  /* Headline Styles */
  --headline1-font-size: 24px;
  --headline1-font-weight: 600;
  --headline1-margin: 0 0 0.5rem;

  --headline2-font-size: 16px;
  --headline2-font-weight: 600;
  --headline2-margin: 1rem 0 0.25rem;

  /* Divider Styles */
  --divider-padding: 0 42px;
  --divider-visibility: visible;

  /* Link Styles */
  --link-text-decoration: none;
  --link-text-decoration-hover: underline;

  /* Input Styles */
  --input-min-width: 14em;

  /* Button Styles */
  --button-min-width: max-content;
}
```

### CSS Shadow Parts

In addition to CSS variables, you can utilize the `::part` selector to customize the styles of various elements.

Please note that shadow parts only function when the web components are attached to the shadow DOM, which is the default
behavior. You can enable the shadow DOM for the components using the following code snippet:

```javascript
register("https://hanko.yourdomain.com", { shadow: true });

// equals

register("https://hanko.yourdomain.com");
```

#### List of all shadow parts

The following parts are available:

- `container` - the UI container
- `headline1` - the "h1" headlines
- `headline2` - the "h2" headlines
- `paragraph` - the paragraph elements
- `button` - every button element
- `primary-button` - the primary button
- `secondary-button` - the secondary button on the email login page
- `input` - every input field
- `text-input` - every input field not used for passcodes
- `passcode-input` - the passcode input fields
- `link` - the links in the footer section
- `error` - the error message container
- `error-text` - the error message
- `divider` - the horizontal divider on the login page
- `divider-text` - the divider text
- `divider-line` - the line before and after the `divider-text`
- `form-item` - the container of a form item, e.g. an input field or a button

#### Examples

The following examples demonstrate how to apply styles to specific shadow parts:

##### Example 1

This example demonstrates how to force the input fields and buttons within the `hanko-auth` component to stack
vertically. The `::part(form-item)` selector targets the `form-item` shadow part within the hanko-auth component, which
is applied using the tag name.

```html
<style>
  hanko-auth::part(form-item) {
    /* Force the input fields and buttons are on top of each other */
    min-width: 100%;
  }
</style>

<hanko-auth></hanko-auth>
```

##### Example 2

This example shows how to adjust the main headlines for all hanko components by targeting the `headline1` shadow part.
The `.hankoComponent::part(headline1)` selector applies the styles to the `headline1` shadow part within elements that
have the `hankoComponent` class.

```html
<style>
  .hankoComponent::part(headline1) {
    /* Adjust the main headlines for all hanko components */
    font-size: 1.3em;
    font-weight: 400;
  }
</style>

<hanko-auth class="hankoComponent"></hanko-auth>
<hanko-profile class="hankoComponent"></hanko-profile>
```

##### Example 3

In this example, a box shadow is applied to the `button` shadow part of the `hanko-auth` component when hovering over
it. The `#hankoAuth::part(button):hover` selector targets the button shadow part within the hanko-auth component using
the ID selector `#hankoAuth` and applies the styles when the `:hover` pseudo-class is active.

```html
<style>
  #hankoAuth::part(button):hover {
    box-shadow: 3px 3px 2px #888;
  }
</style>

<hanko-auth id="hankoAuth"></hanko-auth>
```

### CSS Classes (not recommended)

There is also the possibility to provide your own CSS rules when the web component has not been attached to the shadow
DOM:

```typescript
register("https://hanko.yourdomain.com", { shadow: false });
```

Please take a look at the [CSS example](https://github.com/teamhanko/hanko/raw/main/frontend/elements/example.css) file
to see
which CSS rules can be used. If you only want to change specific properties you can override the predefined ones. For
example if you like to change the background color, include the following CSS rule:

```css
.hanko_container {
  background-color: blue !important;
}
```

Also, you can prevent injecting any styles:

```typescript
register("https://hanko.yourdomain.com", {
  shadow: false,
  injectStyles: false,
});
```

so you don't need to override properties but provide the entirety of CSS rules:

```css
.hanko_container {
  background-color: blue;
}

/* more css rules... */
```

If this is your preferred approach, start with
the [CSS example](https://github.com/teamhanko/hanko/raw/main/frontend/elements/example.css)
file, change everything according to your needs and include the CSS in your page.

Keep in mind we made CSS classes available and added light DOM support only because a Safari bug is breaking the
autocompletion of input elements while the web component is attached to the shadow DOM. You would normally prefer to
attach the component to the shadow DOM and make use of CSS parts for UI customization when the CSS variables are not
sufficient.

## Translations

### Default Behavior

The `hanko-elements` package includes English translations by default and the `lang` attribute can be omitted.

Script:

```typescript
register("https://hanko.yourdomain.com");
```

Markup:

```html
<hanko-auth></hanko-auth>
```

### Installing Additional Translations

Translations are currently available for the following languages:

- "de" - German
- "en" - English
- "fr" - French

You can import them individually:

```typescript
// Replace the paths below with
// "https://cdn.jsdelivr.net/npm/@teamhanko/hanko-elements/dist/i18n/{en|de|all|...}.js"
// if you're using CDN.

import { de } from "@teamhanko/hanko-elements/i18n/de";
import { en } from "@teamhanko/hanko-elements/i18n/en";
import { fr } from "@teamhanko/hanko-elements/i18n/fr";
import { zh } from "@teamhanko/hanko-elements/i18n/zh";
```

Or import all translations at once:

```typescript
import { all } from "@teamhanko/hanko-elements/i18n/all";
```

After importing, provide the translations through the `register()` function:

```typescript
register("https://hanko.yourdomain.com", { translations: { de, en, fr, zh } });

// or

register("https://hanko.yourdomain.com", { translations: all });
```

You can now set the `lang` attribute of the element to the desired language:

```html
<hanko-auth lang="de"></hanko-auth>
```

### Modifying Translations

You can modify existing translations as follows:

```typescript
import { en } from "@teamhanko/hanko-elements/i18n/en";

en.errors.somethingWentWrong = "Aww, snap!";

register("https://hanko.yourdomain.com", { translations: { en } });
```

### Adding New Translations

If you need to create a new translation, pass an object that implements (or partially implements) the `Translation`
interface.

Script:

```typescript
import { all } from "@teamhanko/hanko-elements/i18n/all";
import { Translation } from "@teamhanko/hanko-elements"; // if you're using typescript

const myLang: Translation = {...}

register("https://hanko.yourdomain.com", {translations: {...all, myLang}});
```

Markup:

```html
<hanko-auth lang="myLang"></hanko-auth>
```

### Using External Files

For languages provided via the element's `lang` attribute, or via the [fallback language](#fallback-language) option,
that are not included in the object passed to the `translations` option, the component will fetch a JSON file from the
location specified by the `translationsLocation` option. For example, if "en" is missing due to an empty object being
passed, as shown in the example below, the component will fetch a file named "/i18n/en.json".

Script:

```typescript
register("https://hanko.yourdomain.com", {
  translations: {}, // An empty object, so even the default "en" translation won't be available.
  translationsLocation: "/i18n", // A public folder containing language files, e.g., "en.json".
});
```

Markup:

```html
<!-- Will fetch "/i18n/en.json" -->
<hanko-auth lang="en"></hanko-auth>
```

### Fallback Language

The `fallbackLanguage` option is used to specify a fallback language for the web components when translations are
missing or incomplete for a particular language. By setting the `fallbackLanguage` option to a valid language string
like "en" or "de", the missing translation strings will be automatically retrieved from the specified fallback language.
When the translation for the specified `fallbackLanguage` is not available in the `translations` option, the
web components will attempt to fetch it from an [external file](#using-external-files).

Script:

```typescript
import { en } from "@teamhanko/hanko-elements/i18n/en";
import { Translation } from "@teamhanko/hanko-elements";

const symbols: Partial<Translation> = {
  labels: { continue: "➔" },
};

register("https://hanko.yourdomain.com", {
  fallbackLanguage: "en",
  translations: { en, symbols },
});
```

Markup:

```html
<!-- Will appear in English, but the "continue" button label will be "➔"  -->
<hanko-auth lang="symbols"></hanko-auth>
```

## Experimental Features

### Conditional Mediation / Autofill assisted Requests

```html
<hanko-auth [...] experimental="conditionalMediation" />
```

If the browser supports autofill assisted requests, it will hide the "Sign in with passkey" button on the login page and
instead present the available passkeys via the email input's autocompletion menu. Enabling this feature will currently
cause the following issues:

- On iOS 16/Safari you may encounter an issue that WebAuthn credential registration is not working the first time you
  press the button or only after reloading the page.

- Microsoft Edge v. 108 sometimes crashes or is not able to display the credential name properly.

## Live Demo

Take a look at our [live demo](https://example.hanko.io).

## Examples

The following example implementations are currently available, demonstrating integration into both vanilla JavaScript
and frontend framework environments:

- [A single HTML file](https://raw.githubusercontent.com/teamhanko/hanko/main/frontend/elements/src/example.html) that
  implements most of the features mentioned on this page, with all the key details explained in the comments. You can host
  it on any HTTP server, including locally.
- [Todo apps](https://github.com/teamhanko/hanko/blob/main/frontend/examples/README.md) that demonstrate how integration
  works in various frontend frameworks and provide insights on managing backend communication and JWT validation.

## Frontend framework integrations

To learn more about how to integrate the Hanko elements into frontend frameworks, see our
[guides](https://docs.hanko.io/guides/frontend) in the official documentation and our
[example applications](https://github.com/teamhanko/hanko/blob/main/frontend/examples/README.md).

## Exports

The `@teamhanko/hanko-elements` package exports the functions and interfaces listed below and additionally every
declaration provided by the [frontend-sdk](https://docs.hanko.io/jsdoc/hanko-frontend-sdk/).

### Functions

- `register` - A function to register the web components with the browser's custom element registry.

### Interfaces

- `RegisterOptions` - represents the options of the `register()` function.
- `RegisterResult` - represents the return value of the `register()` function.
- `Translation` - represents a translation that can be provided through the `RegisterOptions`.
- `HankoAuthElementProps` - represents the `<hanko-auth>` element properties.
- `HankoProfileElementProps` - represents the `<hanko-profile>` element properties.
- `HankoEventsElementProps` - represents the `<hanko-events>` element properties.

## Browser support

- Safari
- Firefox
- Opera
- Chromium-based browsers (Chrome, Edge, Brave,...)

## Bugs

- Customizable UI: In Chrome the `::part` selector is not working in combination with some pseudo classes.
  E.g. `:disabled` is currently broken. See:
  [chromium-issue-#1131396](https://bugs.chromium.org/p/chromium/issues/detail?id=1131396),
  [chromium-issue-#953648](https://bugs.chromium.org/p/chromium/issues/detail?id=953648)

Found a bug? Please report on our [GitHub](https://github.com/teamhanko/hanko/issues) page.

## License

The `elements` project is licensed under the [MIT License](LICENSE).
