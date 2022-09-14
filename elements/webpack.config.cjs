const path = require("path");

module.exports = {
  entry: './src/index.ts',
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
      '.js',
      '.sass',
      "declarations.d.ts"
    ],
  },
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'element.hanko-auth.js',
    library: {
      name: 'Hanko',
      type: 'var'
    },
  },
};
