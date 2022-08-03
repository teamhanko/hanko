const path = require("path");

var hankoAuthConfig = {
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      },
      {
        test: /\.(sass)$/,
        use: [
          {
            loader: "style-loader",
            options: {
              injectType: 'singletonStyleTag',
              insert: (styleTag) => {
                // eslint-disable-next-line no-underscore-dangle
                window._hankoStyle = styleTag;
              },
            },
          },
          {
            loader: 'css-loader',
            options: {
              importLoaders: 1,
              modules: true,
            }
          },
          {
            loader: "sass-loader",
            options: {
              sourceMap: true,
            }
          }
        ]
      },
    ]
  },
  resolve: {
    extensions: [
      '.ts',
      '.tsx',
      '.sass'
    ],
  },
  entry: {
    "element.hanko-auth": ['./src/ui/HankoAuth']
  },
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'element.hanko-auth.js',
    library: "HankoAuth",
    libraryTarget: 'umd',
  },
};

var hankoClientConfig = {
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      }]
  },
  resolve: {
    extensions: [
      '.ts',
    ],
  },
  entry: {
    "client": ['./src/lib/Client'],
  },
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'hanko-client.js',
    libraryTarget: 'umd',
    library: "Hanko"
  },
};

module.exports = [hankoAuthConfig, hankoClientConfig];
