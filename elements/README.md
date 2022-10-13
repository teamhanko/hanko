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

The web component needs to be registered first. You can control whether it should be attached to the shadow DOM or not
using the `shadow` property. It's set to true by default, and you will be able to use the CSS parts
to change the appearance of the component.

There is currently an issue with Safari browsers, which breaks the autocompletion feature of
input fields when the component is shadow DOM attached. So if you want to make use of the conditional UI or other
autocompletion features you must set `shadow` to false. The disadvantage is that the CSS parts are not working anymore, and you must
style the component by providing your own CSS properties. CSS variables will work in both cases.

Use as a module:

```typescript
import { register } from "@teamhanko/hanko-elements/hanko-auth"

register({
  shadow: true,      // Set to false if you don't want the web component to be attached to the shadow DOM.
  injectStyles: true // Set to false if you don't want to inject any default styles.
})
```

With a script tag via CDN:

```html
<script src="https://cdn.jsdelivr.net/npm/@teamhanko/hanko-elements/dist/element.hanko-auth.js"/>

<script>
    HankoAuth.register({ shadow: true, injectStyles: true })
</script>
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

These events bubble up through the DOM tree.

- `hankoAuthSuccess` - Login or registration completed successfully and a JWT has been issued. You can now take control and redirect the user to protected pages.

```js
document.addEventListener('hankoAuthSuccess', () => {
    document.body.innerHTML = 'secured content...'
})
```

## Demo

The animation below demonstrates how user registration with passwords enabled looks like. You can set up the flow you
like using the [Hanko API](https://github.com/teamhanko/hanko/blob/main/backend/README.md) configuration file. The registration flow also includes email
verification via passcodes and the registration of a passkey so that the user can log in without passwords or passcodes.

<img src="https://github.com/teamhanko/hanko/raw/main/elements/demo.gif" width="410"/>

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

  --container-padding: 20px;
  --container-max-width: 600px;

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

### CSS classes

There is also the possibility to provide your own CSS rules when the web component has not been attached to the shadow
DOM:

```typescript
register({ shadow: false })
```

Please take a look at the [CSS example](https://github.com/teamhanko/hanko/raw/main/elements/example.css) file to see
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

If this is your preferred approach, start with the [CSS example](https://github.com/teamhanko/hanko/raw/main/elements/example.css)
file, change everything according to your needs and include the CSS in your page.

Keep in mind we made CSS classes available and added light DOM support only because a Safari bug is breaking the
autocompletion of input elements while the web component is attached to the shadow DOM. You would normally prefer to
attach the component to the shadow DOM and make use of CSS parts for UI customization when the CSS variables are not
sufficient.

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

  --container-max-width: 450px;
  --container-padding: 10px 20px;

  --headline-font-weight: 800;
  --headline-font-size: 24px;

  --lightness-adjust-dark: 30%;
  --lightness-adjust-dark-light: 10%;
  --lightness-adjust-light: -10%;
  --lightness-adjust-light-dark: 30%;
}

hanko-auth::part(headline),
hanko-auth::part(input),
hanko-auth::part(link) {
  color: hsl(33, 93%, 55%);
}

hanko-auth::part(link):hover {
  text-decoration: underline;
}

hanko-auth::part(button):hover,
hanko-auth::part(input):focus {
  border-width: 2px;
}
```

Result:

<img src="https://github.com/teamhanko/hanko/raw/main/elements/demo-ui.png" width="450"/>

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

