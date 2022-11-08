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
  ]
};
