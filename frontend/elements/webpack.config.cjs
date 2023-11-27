const path = require("path");

module.exports = {
  experiments: {
    outputModule: true,
  },
  entry: {
    hankoElements: {
      filename: "elements.js",
      import: "./src/index.ts",
      library: {
        type: "module",
      },
    },
    de: {
      filename: "i18n/de.js",
      import: "./src/i18n/de.ts",
      library: {
        type: "module",
      },
    },
    en: {
      filename: "i18n/en.js",
      import: "./src/i18n/en.ts",
      library: {
        type: "module",
      },
    },
    fr: {
      filename: "i18n/fr.js",
      import: "./src/i18n/fr.ts",
      library: {
        type: "module",
      },
    },
    fr: {
      filename: "i18n/it.js",
      import: "./src/i18n/it.ts",
      library: {
        type: "module",
      },
    },
    ptBR: {
      filename: "i18n/pt-BR.js",
      import: "./src/i18n/pt-BR.ts",
      library: {
        type: "module",
      },
    },
    zr: {
      filename: "i18n/zh.js",
      import: "./src/i18n/zh.ts",
      library: {
        type: "module",
      },
    },
    bn: {
      filename: "i18n/bn.js",
      import: "./src/i18n/bn.ts",
      library: {
        type: "module",
      },
    },
    all: {
      filename: "i18n/all.js",
      import: "./src/i18n/all.ts",
      library: {
        type: "module",
      },
    },
  },
  module: {
    rules: [
      {
        test: /\.(tsx?)$/,
        use: "ts-loader",
        exclude: [/node_modules/, /dist/],
        resolve: {
          fullySpecified: false,
        },
      },
      {
        test: /\.(sass)$/,
        use: [
          {
            loader: "style-loader",
            options: {
              injectType: "singletonStyleTag",
              insert: (styleTag) => {
                // eslint-disable-next-line no-underscore-dangle
                window._hankoStyle = styleTag;
              },
            },
          },
          {
            loader: "css-loader",
            options: {
              modules: {
                localIdentName: "hanko_[local]",
                localIdentContext: path.resolve(__dirname, "src"),
              },
              importLoaders: 1,
            },
          },
          {
            loader: "sass-loader",
            options: {
              sourceMap: true,
            },
          },
        ],
      },
    ],
  },
  resolve: {
    extensions: [".ts", ".tsx", ".js", ".sass", "declarations.d.ts"],
  },
  output: {
    clean: true,
    path: path.resolve(__dirname, "dist"),
  },
};
