# Contributing to Hanko

Thank you for considering contributing to Hanko! Following are the guidelines we would like you to follow:

- [Code of Conduct](#code-of-conduct)
- [Communication](#communication)
- [Reporting Issues](#reporting-issues)
  - [Security](#security)
  - [Bugs](#bugs)
- [Feature Requests](#feature-requests)
- [Submitting Code](#submitting-code)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Style Guidelines](#style-guidelines)

## Code of Conduct

We expect all contributors to adhere to our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Communication

If you have any questions, want to discuss bugs or feature requests, or just want talk to other Hanko users you are welcome
to join our [Slack](https://hanko.io/community) community or use the [Hanko Discussions](https://github.com/teamhanko/hanko/discussions)
(especially useful for long term discussion or larger questions).

## Reporting issues

Reporting issues requires a [GitHub](https://github.com/) account. Please do not use the issue
tracker for general support questions but use the above mentioned [communication](#communication) channels.

### Security

Pursuant to our [security policy](./SECURITY.md), any security vulnerabilities should be reported directly to
`security@hanko.io` instead of using the issue tracker.

### Bugs

Bugs are tracked as [GitHub issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/about-issues).
When reporting a bug choose "Bug Report" when [creating](https://github.com/teamhanko/hanko/issues/new/choose) a new
issue. Doing so will present you with a template form to fill out. Be as detailed as possible. Good
bug reports are vital, so thank you for taking the time!

Before reporting a bug:

1. Take a look at the [existing issues](https://github.com/teamhanko/hanko/issues?q=is%3Aissue+label%3Abug) and make
   sure the issue hasn't already been reported. If you find a similar bug and the issue is still open, consider adding
   a comment providing any new information that you might be able to report.
2. Make sure the issue hasn't been fixed already. Try to reproduce it using the latest `main` branch in the repository if
   you were not working with the latest version.
3. Make sure the bug is really a bug. If you need general support or are unsure whether some behaviour represents a bug
   please do not file a bug ticket but reach out through the above mentioned [communication](#communication) channels.
4. Gather as much information about the bug as you can. Logs, screenshots/screen captures, steps to reproduce the bug
   can be vital for a useful bug report.

If you already have suggestions on how to fix the bug, do not hesitate to include them in the bug description.

## Feature requests

Just like bugs, feature requests are tracked as [GitHub issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/about-issues).
When suggesting an enhancement choose "Feature Request" when [creating](https://github.com/teamhanko/hanko/issues/new/choose) a new
issue and fill out the template form.

Before making a feature request:

1. Take a look at the [existing issues](https://github.com/teamhanko/hanko/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement) and make
   sure the issue hasn't already been reported.
2. Make sure the issue hasn't been implemented already. Always try using the latest `main` branch in the repository if
   to confirm the feature is not already in place.

When filling out the template form, be sure to be as detailed as possible. Describe the current behavior and explain
which behavior you expect to see instead. Explain why this enhancement would be useful to Hanko users.

## Submitting Code

Contributing code requires a [GitHub](https://github.com/) account. All contributions are made via
[pull requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests).
Pull requests should target the `main` branch.

To submit your code:

1. [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) the [repository](https://github.com/teamhanko/hanko).
2. [Clone](https://docs.github.com/en/get-started/quickstart/fork-a-repo#cloning-your-forked-repository) the forked repository.
3. [Configure remotes](https://docs.github.com/en/get-started/quickstart/fork-a-repo#configuring-git-to-sync-your-fork-with-the-original-repository).
4. Create a new [branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches)
   off of the `main` branch.
   ```
   git checkout -b <new-branch-name>
   ```
5. Make your changes. Make sure to follow the [Style Guidelines](#style-guidelines). Commit your changes.
   ```
   git add -A
   git commit
   ```
   Commit messages should follow the [Commit Message Guidelines](#commit-message-guidelines).
6. Make sure to update, or add to any tests where appropriate. Try to run tests locally first (`go test ./...` for the
   `backend`, see the [README](./e2e/README.md) for the `e2e`tests on how to run them).
7. If you added or changed a feature, make sure to document it in the README.md file. If your change
   affects the `backend` API update the [Open API spec(s)](./docs/static/spec).
   If your changes affect the `backend` configuration, update the [Config.md](./backend/docs/Config.md).
8. Push your feature branch up to your fork:
   ```
   git push origin <feature-branch-name>
   ```
9. [Create a pull request from your fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork).
10. Submit the pull request by filling out the [pull request template](./.github/PULL_REQUEST_TEMPLATE.md)
    (note: the template should be displayed automatically once you open a pull request; take account of the comments in
    the displayed template).
11. If a pull request is not ready to be reviewed it should be marked as a "Draft".


When pull requests fail test checks, authors are expected to update
their pull requests to address the failures until the tests pass. If you have trouble or questions on how to add to
existing tests, reach out through our [communication](#communication) channels.

# Commit Message Guidelines

Commit messages should adhere to the
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
The commit message should be structured as follows:

```
<type>(<optional scope>): <description>

<optional body>

<optional footer(s)>
```

The commit message headline should have the following structure:
```
<type>(<optional scope>): <description>
   │            │               │
   │            │               └─⫸ Summary in present tense. Not capitalized. No period at the end.
   │            │
   │            └─⫸ Commit Scope: optional
   │
   └─⫸ Commit Type: build|ci|docs|feat|fix|perf|refactor|test
```
The `<type>` should be one of the following:
* **build**: Changes that affect the build system or external dependencies
* **ci**: Changes that affect the CI workflows (e.g. changes to `.github` CI configuration files)
* **docs**: Documentation only changes (this includes both content in the `docs` as well as changes to readmes)
* **feat**: A new feature
* **fix**: A bug fix
* **perf**: A code change that improves performance
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **test**: Adding missing tests or correcting existing tests

The `<scope>` is optional. If present, it should be the name of the (npm) package or directory affected by the changes of
the commit.

# Style Guidelines

## Go

Go files should be [formatted](https://go.dev/blog/gofmt) according to gofmt's rules.

```
# single file
go fmt path/to/changed/file.go

# all files, e.g. in 'backend' directory
go fmt ./...
```
