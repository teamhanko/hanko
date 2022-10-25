# Hanko Documentation

The Hanko documentation is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

### Installation

```
$ npm install
```

### Local Development

```
$ npm run start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```
$ npm run build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Swizzled components

The following components have been [swizzled](https://docusaurus.io/docs/swizzling):

### `DocCard`

- Make icons used in the DocCard customizable via `sidebar_custom_props`
- Make description toggle-able via `sidebar_custom_props`

Uses the following `sidebar_custom_props`:

| Name                   | Type    | Default         | Description                                                                                                                                                                                                                    |
|------------------------|---------|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| docCardIconName        | string  | undefined       | Name of the icon file without file extension. Icons must be placed in `static/img/icons` folder and have an *.svg extension. Uses default images if `undefined`.                                                               |
| docCardShowDescription | boolean | undefined/false | Whether to show the description in auto-generated link/category cards. Introduced for SEO reasons such that front-matter `description`s can be used to generate meta tags without being forced to display them in the DocCard. |
