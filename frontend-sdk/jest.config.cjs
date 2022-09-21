/** @type {import('ts-jest/dist/types').InitialOptionsTsJest} */
module.exports = {
  globals: {
    "ts-jest": {
      "tsconfig": "tsconfig.json"
    }
  },
  preset: "ts-jest",
  testEnvironment: "jsdom",
  coverageProvider: "v8",
};
