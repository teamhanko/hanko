module.exports = {
  'env': {
    'browser': true,
    'es2021': true,
    'node': true,
  },
  'extends': [
    'eslint:recommended',
    'google',
    'preact',
    'plugin:promise/recommended',
    // "plugin:@typescript-eslint/recommended",
    // "plugin:@typescript-eslint/recommended-requiring-type-checking",
    'plugin:prettier/recommended',
  ],
  'parser': '@typescript-eslint/parser',
  'parserOptions': {
    'ecmaVersion': 'latest',
    'sourceType': 'module',
    "project": "tsconfig.json",
    "tsconfigRootDir": ".",
  },
  'plugins': [
    '@typescript-eslint'
  ],
  'rules': {
    'no-unused-vars': ['error', { 'args': 'none' }]
  }
};
