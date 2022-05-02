const path = require("path");
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;
module.exports = {
  entry: './src/index.ts',
  plugins: [
    new BundleAnalyzerPlugin()
  ],
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
    extensions: ['.ts', '.tsx', '.js', '.css', "declarations.d.ts"],
  },
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'hanko.js',
    library: {
      name: 'Hanko', type: 'var'
    },
  },
};
