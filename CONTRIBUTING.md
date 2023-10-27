# Contributing to Hanko

Thank you for considering contributing to Hanko! We have provided the following guidelines to ensure a smooth and effective contribution process.

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

We kindly request that all contributors adhere to our [Code of Conduct](./CODE_OF_CONDUCT.md) to maintain a respectful and inclusive enviroment.

## Communication

For questions, bug reports, feature requests, or discussions with other Hanko users, please feel free to participate in our community through the following channels:

- [Slack Community](https://hanko.io/community)
- [Hanko Discussions](https://github.com/teamhanko/hanko/discussions) (ideal for in-depth discussions and larger questions)

## Reporting Issues

To report issues, please ensure you have a [GitHub](https://github.com/) account. For general support inquiries, please utilize the [communication channels](#communication) mentioned above and refrain from using the issue tracker.

### Security

For security-related matters, including vulnerabilities, please follow our [security policy](./SECURITY.md) and report them directly to `security@hanko.io` rather than using the issue tracker.

### Bugs

Bugs are tracked as [GitHub issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/about-issues). When reporting a bug, select "Bug Report" when [creating a new issue](https://github.com/teamhanko/hanko/issues/new/choose). This will present you with a template to complete. Please provide as much detail as possible, as thorough bug reports are crucial.

Before reporting a bug:

1. Review the [existing issues](https://github.com/teamhanko/hanko/issues?q=is%3Aissue+label%3Abug) to ensure it hasn't already been reported. If you find a similar bug with an openissue, consider adding a comment with new information.
2. Confirm the bug has not already been resolved by testing with the latest `main` branch in the repository if you were not working with the latest version.
3. Verify the issue is indeed a bug. If you need general support or are uncertain whether a behavior constitutes a bug, please use the [communication channels](#communication) mentioned above instead.
4. Gather comprehemsive information about the bug, including logs, screenshots, and steps to reproduce the issue. This information is invaluable for effective bug reporting.

If you have suggestions for resolving the bug, please include them in the bug description.

## Feature Requests

Feature requests are also tracked as [GitHub issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/about-issues). To propose an enhancement, select "Feature Request" when [creating a new issue](https://github.com/teamhanko/hanko/issues/new/choose) and complete the provided template.

Before submitting a feature request:

1. Examine the [existing issues](https://github.com/teamhanko/hanko/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement) to ensure it hasn't already been reported.
2. Confirm the feature hasn't already been implemented by testing with the latest `main` branch in the repository.
3. When completing the template, provide detailed informnation, including the current behavior, the desired behavior, and an explanation of how the enhancement would benefit Hanko users.

## Submitting Code

To contribute code, you will need a [GitHub](https://github.com/) account. All contributions are made via [pull requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests), which should target the `main` branch.

To submit your code:

1. [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) the [repository](https://github.com/teamhanko/hanko).
2. [Clone](https://docs.github.com/en/get-started/quickstart/fork-a-repo#cloning-your-forked-repository) the forked repiository.
3. [Configure remotes](https://docs.github.com/en/get-started/quickstart/fork-a-repo#configuring-git-to-sync-your-fork-with-the-original-repository).
4. Create a new [branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches) from the `main` branch.
   ```
   git checkout -b <new-branch-name>
   ```
5. Make your changes. Make sure to follow the [Style Guidelines](#style-guidelines). Commit your changes.
   ```
   git add -A
   git commit
   ```
  Commit messages should follow the [Commit Message Guidelines](#commit-message-guidelines).

6. Make sure to update or add to any tests where appropriate. Attempt to run tests locally first:
   - For the `backend`, use `go test ./...`.
   - For the `e2e` tests, please refer to the [README](./e2e/README.md) for instructions on how to run them.

7. If you have added or modified a feature, it is essential to document it in the README.md file. If your change impacts the `backend` API, ensure that you update the [Open API spec(s)](./docs/static/spec). If your changes affect the `backend` configuration, update the [Config.md](./backend/docs/Config.md).

8. Push your feature branch to your fork:
   ```shell
   git push origin <feature-branch-name>
   ```
9. [Create a pull request from your fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork).

10. After creating the pull request, please submit it by completing the [pull request template](./.github/PULL_REQUEST_TEMPLATE.md). Note that the template should be displayed automatically once you open a pull request. Be sure to consider the comments and instructions provided in the displayed template.

11. If a pull request is not yet ready for review, please mark it as a "Draft."


When pull requests fail test checks, authors are expected to update their pull requests to address the failures until the tests pass. If you encounter difficulties or have questions about how to add to existing tests, please reach out through our [communication](#communication) channels.

# Commit Message Guidelines

Commit messages should follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification. The commit message structure should be as follows:

```
<type>(<optional scope>): <description>

<optional body>

<optional footer(s)>
```

The commit message headline should adhere to the following structure:
```
<type>(<optional scope>): <description>
   │            │               │
   │            │               └─⫸ Summary in present tense. Not capitalized. No period at the end.
   │            │
   │            └─⫸ Commit Scope: optional
   │
   └─⫸ Commit Type: build|ci|docs|feat|fix|perf|refactor|test|chore
```
The `<type>` should be one of the following:
* **build**: Changes that affect the build system
* **ci**: Changes that affectthe CI workflows (e.g., modifications to `.github` CI configuration files)
* **docs**: Documentation-only changes (this includes both content in the `docs` directory and changes to readmes)
* **feat**: Introducing a new feature
* **fix**: Fixing a bug
* **perf**: Code changes aimed at improving performance
* **refactor**: Code changes that neither fix a bug nor add a new feature
* **test**: Adding missing tests or correcting existing tests
* **chore**: Anything that cannot be properly categorized using the above prefixes (e.g., version updates)

The `<scope>` is optional. If present, it should represent the name of the (npm) package or directory impacted by the comit changes.

# Style Guidelines

## Go
Go files should be [formatted](https://go.dev/blog/gofmt) according to gofmt's rules.

```
# single file
go fmt path/to/changed/file.go

# all files, e.g. in 'backend' directory
go fmt ./...
```
