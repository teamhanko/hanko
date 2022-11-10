const path = require("path");

module.exports = {
  entry: {
    hankoAuth: {
      filename: 'element.hanko-auth.js',
      import: './src/index.ts',
      library: {
        name: 'HankoAuth',
        type: 'umd',
        umdNamedDefine: true,
      },
    }
  },
  module: {
    rules: [
      {
        test: /\.(tsx?)$/,
        use: 'ts-loader',
        exclude: [/node_modules/, /dist/],
        resolve: {
          fullySpecified: false
        },
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
              modules: {
                localIdentName: "hanko_[local]",
                localIdentContext: path.resolve(__dirname, "src"),
              },
              importLoaders: 1,
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
      '.js',
      '.sass',
      "declarations.d.ts"
    ],
  },
  output: {
    clean: true,
    path: path.resolve(__dirname, 'dist'),
  },
};
