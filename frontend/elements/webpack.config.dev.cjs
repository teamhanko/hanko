const baseConfig = require("./webpack.config.cjs");

baseConfig.module.rules.push({
  test: /\.c?js$/,
  enforce: "pre",
  use: ["source-map-loader"],
})

module.exports = {
  devtool: 'eval-source-map',
  ...baseConfig
};
