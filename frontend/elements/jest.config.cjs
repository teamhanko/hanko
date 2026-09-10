/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: "ts-jest",
  testEnvironment: "jsdom",
  roots: ["<rootDir>/tests"],
  moduleNameMapper: {
    "\\.(css|sass|scss)$": "<rootDir>/tests/styleMock.cjs",
    "^preact$": "<rootDir>/../node_modules/preact/dist/preact.umd.js",
    "^preact/compat$":
      "<rootDir>/../node_modules/preact/compat/dist/compat.umd.js",
    "^preact/hooks$":
      "<rootDir>/../node_modules/preact/hooks/dist/hooks.umd.js",
    "^preact/jsx-runtime$":
      "<rootDir>/../node_modules/preact/jsx-runtime/dist/jsxRuntime.umd.js",
    "^preact/test-utils$":
      "<rootDir>/../node_modules/preact/test-utils/dist/testUtils.umd.js",
  },
  transform: {
    "^.+\\.tsx?$": [
      "ts-jest",
      {
        tsconfig: "<rootDir>/tsconfig.test.json",
      },
    ],
  },
  coverageProvider: "v8",
};
