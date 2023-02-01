# Hanko elements

Provides web components that will bring a modern login and registration experience
to your users. It integrates the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md), a backend
that provides the underlying functionalities.

## Features

* Registration and login flows with and without passwords
* Passkey authentication
* Passcodes, a convenient way to recover passwords and verify email addresses
* Email, Password and Passkey management
* Customizable UI

## Installation

```shell
# npm
npm install @teamhanko/hanko-elements

# yarn
yarn add @teamhanko/hanko-elements

# pnpm
pnpm install @teamhanko/hanko-elements
```

## Usage

### Script

The web components need to be registered first. You can control whether they should be attached to the shadow DOM or not
using the `shadow` property. It's set to true by default, and it's possible to make use of the [CSS shadow parts](#css-shadow-parts)
to change the appearance of the component. [CSS variables](#css-variables) will work in both cases.

Use as a module:

```typescript
import { register } from "@teamhanko/hanko-elements"

register({
  shadow: true,      // Set to false if you don't want the web component to be attached to the shadow DOM.
  injectStyles: true // Set to false if you don't want to inject any default styles.
})
```

With a script tag via CDN:

```html
<script src="https://cdn.jsdelivr.net/npm/@teamhanko/hanko-elements/dist/elements.js"/>

<script>
  Elements.register({shadow: true, injectStyles: true})
</script>
```

### &lt;hanko-auth&gt;

A web component that handles user login and user registration.

#### Markup

```html
<hanko-auth api="https://hanko.yourdomain.com" lang="en"/>
```

#### Attributes

- `api` the location where the Hanko API is running.
- `lang` Currently supported values are "en" for English and "de" for German. If the value is omitted, "en" is used.
- `experimental` A space-seperated list of experimental features to be enabled. See [experimental features](#experimental-features).

#### Events

These events bubble up through the DOM tree.

- `hankoAuthSuccess` - Login or registration completed successfully and a JWT has been issued. You can now take control and redirect the user to protected pages.

```js
document.addEventListener('hankoAuthSuccess', () => {
    document.body.innerHTML = 'secured content...'
})
```

### &lt;hanko-profile&gt;

A web component that allows to manage emails, passwords and passkeys.

#### Markup

```html
<hanko-profile api="https://hanko.yourdomain.com" lang="en"/>
```

#### Attributes

- `api` the location where the Hanko API is running.
- `lang` Currently supported values are "en" for English and "de" for German. If the value is omitted, "en" is used.

## UI Customization

### CSS Variables

CSS variables can be used to style the `hanko-auth` and  `hanko-profile` elements to your needs. A list of all CSS
variables including default values can be found below:

```css
hanko-auth, hanko-profile {
  /* Color Scheme */
  --color: #171717
  --color-shade-1: #8f9095
  --color-shade-2: #e5e6ef

  --brand-color: #506cf0
  --brand-color-shade-1: #6b84fb
  --brand-contrast-color: white

  --background-color: white
  --error-color: #e82020
  --link-color: #506cf0

  /* Font Styles */
  --font-weight: 400
  --font-size: 14px
  --font-family: sans-serif

  /* Border Styles */
  --border-radius: 4px
  --border-style: solid
  --border-width: 1px

  /* Item Styles */
  --item-height: 34px
  --item-margin: .5rem 0

  /* Container Styles */
  --container-padding: 0
  --container-max-width: 600px

  /* Headline Styles */
  --headline1-font-size: 24px
  --headline1-font-weight: 600
  --headline1-margin: 0 0 .5rem

  --headline2-font-size: 14px
  --headline2-font-weight: 600
  --headline2-margin: 1rem 0 .25rem

  /* Divider Styles */
  --divider-padding: 0 42px
  --divider-display: block
  --divider-visibility: visible

  /* Link Styles */
  --link-text-decoration: none
  --link-text-decoration-hover: underline

  /* Input Styles */
  --input-min-width: 12em

  /* Button Styles */
  --button-min-width: max-content
}
```

### CSS Shadow Parts

In addition to the CSS variables, there is the possibility of using the `::part` selector to equip various elements
with your own styles.

The following parts are available:

- `container` - the UI container
- `headline` - the headline of each page
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
- `form-item` - the container of a form item, e.g. an input field or a button

### CSS classes

There is also the possibility to provide your own CSS rules when the web component has not been attached to the shadow
DOM:

```typescript
register({ shadow: false })
```

Please take a look at the [CSS example](https://github.com/teamhanko/hanko/raw/main/frontend/elements/example.css) file to see
which CSS rules can be used. If you only want to change specific properties you can override the predefined ones. For
example if you like to change the background color, include the following CSS rule:

```css
.hanko_container {
  background-color: blue !important;
}
```

Also, you can prevent injecting any styles:

```typescript
register({ shadow: false, injectStyles: false })
```

so you don't need to override properties but provide the entirety of CSS rules:

```css
.hanko_container {
  background-color: blue;
}

/* more css rules... */
```

If this is your preferred approach, start with the [CSS example](https://github.com/teamhanko/hanko/raw/main/frontend/elements/example.css)
file, change everything according to your needs and include the CSS in your page.

Keep in mind we made CSS classes available and added light DOM support only because a Safari bug is breaking the
autocompletion of input elements while the web component is attached to the shadow DOM. You would normally prefer to
attach the component to the shadow DOM and make use of CSS parts for UI customization when the CSS variables are not
sufficient.

## Experimental Features

### Conditional Mediation / Autofill assisted Requests

```html
<hanko-auth [...] experimental="conditionalMediation"/>
```

If the browser supports autofill assisted requests, it will hide the "Sign in with passkey" button on the login page and
instead present the available passkeys via the email input's autocompletion menu. Enabling this feature will currently
cause the following issues:

- On iOS 16/Safari you may encounter an issue that WebAuthn credential registration is not working the first time you
  press the button or only after reloading the page.

- Microsoft Edge v. 108 sometimes crashes or is not able to display the credential name properly.

## Demo

Take a look at our [live demo](https://example.hanko.io).

## Frontend framework integrations

To learn more about how to integrate the Hanko elements into frontend frameworks, see our
[guides](https://docs.hanko.io/guides/frontend) in the official documentation and our
[example applications](https://github.com/teamhanko/hanko/blob/main/examples/README.md).

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

