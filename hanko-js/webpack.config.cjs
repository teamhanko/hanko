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
        test: /\.css$/,
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
      '.css',
      "declarations.d.ts"
    ],
  },
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'element.hanko-ui.js',
    library: {
      name: 'Hanko',
      type: 'var'
    },
  },
};
