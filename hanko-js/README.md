# &lt;hanko-auth&gt; element

The `<hanko-auth>` element offers a complete user interface that will bring a modern login and registration experience
to your users. It integrates the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md), a backend
that provides the underlying functionalities.

## Features

* Registration and login flows with and without passwords
* Passkey authentication
* Passcodes, a convenient way to recover passwords and verify email addresses
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

Import as a module:

```typescript
import "@teamhanko/hanko-elements/hanko-auth"
```

With a script tag via CDN:

```html
<script type="module" src="https://unpkg.com/hanko-elements/dist/element.hanko-auth.js">
```

### Markup

```html
<hanko-auth api="https://hanko.yourdomain.com" lang="en"/>
```

Please take a look at the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md) to see how to spin up the backend.

Note that we're working on Hanko Cloud, so that you don't need to run the Hanko API by yourself and all you need is to
do is adding the `<hanko-auth>` element to your page.

## Attributes

- `api` the location where the Hanko API is running.
- `lang` Currently supported values are "en" for English and "de" for German. If the value is omitted, "en" is used.

## Events

Events are dispatched on the `<hanko-auth>` element. These events do not bubble.

- `success` - Login or registration completed successfully and a JWT has been issued. You can now take control and
  redirect the user to protected pages.

```js
const hanko = document.querySelector('hanko-auth')

hanko.addEventListener('success', () => {
    hanko.parentElement.innerHTML = 'secured content...'
})
```

## Demo

The animation below demonstrates how user registration with passwords enabled looks like. You can set up the flow you
like using the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md) configuration file. The registration flow also includes email
verification via passcodes and the registration of a passkey so that the user can log in without passwords or passcodes.

<img src="https://github.com/teamhanko/hanko/raw/main/hanko-js/demo.gif" width="410"/>

## UI Customization

### CSS Variables

CSS variables can be used to style the `hanko-auth` element to your needs. Based on preset values and provided CSS
variables, individual elements will be styled, including color shading for different UI states (e.g. hover, focus,..).

Note that colors must be provided as individual HSL values. We'll have to be patient, unfortunately, until
broader browser support for relative colors arrives, which would allow native CSS colors to be used.

A list of all CSS variables including default values can be found below:

```css
hanko-auth {
  --background-color-h: 0;
  --background-color-s: 0%;
  --background-color-l: 100%;

  --border-radius: 3px;
  --border-style: solid;
  --border-width: 1.5px;

  --brand-color-h: 351;
  --brand-color-s: 100%;
  --brand-color-l: 59%;

  --color-h: 0;
  --color-s: 0%;
  --color-l: 0%;

  --container-max-height: 500px;
  --container-max-width: 450px;
  --container-min-height: 500px;
  --container-min-width: 450px;
  --container-padding: 25px;

  --error-color-h: 351;
  --error-color-s: 100%;
  --error-color-l: 59%;

  --font-family: sans-serif;
  --font-size: 16px;
  --font-weight: 400;

  --headline-font-size: 30px;
  --headline-font-weight: 700;

  --input-height: 50px;

  --item-margin: 15px 0;

  --lightness-adjust-dark: -30%;
  --lightness-adjust-dark-light: -10%;
  --lightness-adjust-light: 10%;
  --lightness-adjust-light-dark: 30%;
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

### Example

The example below shows how you can use CSS variables in combination with styled shadow DOM parts:

```css
hanko-auth {
  --color-h: 188;
  --color-s: 99%;
  --color-l: 38%;

  --brand-color-h: 315;
  --brand-color-s: 100%;
  --brand-color-l: 59%;

  --background-color-h: 196;
  --background-color-s: 10%;
  --background-color-l: 21%;

  --border-width: 1px;
  --border-radius: 5px;

  --font-weight: 400;
  --font-size: 16px;
  --font-family: Helvetica;

  --input-height: 45px;
  --item-margin: 10px;

  --container-min-height: 0;
  --container-min-width: 350px;
  --container-max-width: 450px;
  --container-padding: 10px 20px;

  --headline-font-weight: 800;
  --headline-font-size: 24px;

  --lightness-adjust-dark: 30%;
  --lightness-adjust-dark-light: 10%;
  --lightness-adjust-light: -10%;
  --lightness-adjust-light-dark: 30%;
}

hanko-auth::part(headline) {
  color: hsl(33, 93%, 55%);
}

hanko-auth::part(input) {
  color: hsl(33, 93%, 55%);
}

hanko-auth::part(input-text)::first-line {
  color: hsl(33, 93%, 45%);
}

hanko-auth::part(link) {
  color: hsl(33, 93%, 55%);
}

hanko-auth::part(link):hover {
  text-decoration: underline;
}
```

Result:

<img src="https://github.com/teamhanko/hanko/raw/main/hanko-js/demo-ui.png" width="450"/>

## Browser support

- Chrome
- Firefox
- Safari
- Microsoft Edge

## Bugs

- Customizable UI: In Chrome the `::part` selector is not working in combination with some pseudo classes.
E.g. `:disabled` is currently broken. See:
[chromium-issue-#1131396](https://bugs.chromium.org/p/chromium/issues/detail?id=1131396),
[chromium-issue-#953648](https://bugs.chromium.org/p/chromium/issues/detail?id=953648)

Found a bug? Please report on our [GitHub](https://github.com/teamhanko/hanko/issues) page.

## License

The `hanko-js` project is licensed under the [MIT License](LICENSE).

