// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Hanko Documentation',
  tagline: 'Hanko Documentation',
  url: 'https://docs.hanko.io',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  scripts: [{ src: 'https://plausible.io/js/plausible.js', defer: true, 'data-domain': 'docs.hanko.io' }],

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'teamhanko', // Usually your GitHub org/user name.
  projectName: 'docs', // Usually your repo name.
  trailingSlash: false,

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'redocusaurus',
      {
        // Plugin Options for loading OpenAPI files
        specs: [
          {
            spec: 'static/spec/public.yaml',
            route: '/api/public',
          },
          {
            spec: 'static/spec/admin.yaml',
            route: '/api/admin',
          },
        ],
        // Theme Options for modifying how redoc renders them
        theme: {
          primaryColor: '#ff2e4c',
          primaryColorDark: '#aedfff',
          options : {
            nativeScrollbars: true,
            scrollYOffset: 60,
            disableSearch: true,
            noAutoAuth: true,
            expandResponses: '200',
            jsonSampleExpandLevel: 3,
            pathInMiddlePanel: false,
            requiredPropsFirst: true,
            hideDownloadButton: true
          },
          // Change with your site colors
          theme: {
            typography: {
              fontSize: '16px',
              fontWeightRegular: '500',
              code: {
                fontSize: '13px',
                fontFamily: 'Courier, monospace',
              }
            },
            sidebar: {
              arrow: {
                size: '1.7em',
                color: '#7f7f7f'
              },
              width: '300px',
            },
          }
        },
      },
    ],
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/', // Serve the docs at the site's root
          sidebarPath: require.resolve('./sidebars.js'),
          remarkPlugins: [require('@docusaurus/remark-plugin-npm2yarn'), {sync: true}],
        },
        blog: false,
        theme: {
          customCss: [require.resolve('./src/css/custom.css'), require.resolve('./src/css/redoc.css')]
        },
      }),
    ],
  ],

  themeConfig:
  /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      docs: {
        sidebar: {
          hideable: true
        }
      },
      image: 'img/thumbnail.jpg',
      colorMode: {
        defaultMode: 'dark',
        disableSwitch: true,
      },
      navbar: {
        logo: {
          alt: 'Hanko Logo',
          src: 'img/logo.svg',
          href: '/',
        },
        items: [
          {
            to: '/',
            label: 'Docs',
            position: 'left',
            activeBaseRegex: '^((?!\/api).)*$'
          },
          {
            type: 'dropdown',
            label: 'API',
            position: 'left',
            items: [
              {
                label: 'Public',
                to: 'api/public'
              },
              {
                label: 'Admin',
                to: 'api/admin'
              }
            ]
          },
          {
            to: 'jsdoc/hanko-frontend-sdk',
            label: 'SDK',
            position: 'left',
            target: '_blank'
          },
          {
            href: 'https://github.com/teamhanko/hanko',
            title: "Visit us on GitHub!",
            position: 'right',
            className: 'header-github-link',
            'aria-label': 'Visit us on GitHub!',
          },
        ],
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['php', 'bash']
      },
    }),
};

module.exports = config;
