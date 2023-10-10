# Contributing to Hanko

Welcome to the Hanko community! We're thrilled that you're considering contributing. Below, we've laid out some guidelines to make your journey smooth and enjoyable.

## Code of Conduct

First things first, please familiarize yourself with our [Code of Conduct](./CODE_OF_CONDUCT.md). It's a quick read but sets the tone for a positive and inclusive community.

## Communication

Got questions or just want to connect with other Hanko enthusiasts? Join us on [Slack](https://hanko.io/community) or dive into [Hanko Discussions](https://github.com/teamhanko/hanko/discussions). We're here to help and chat!

## Reporting Issues

If you encounter issues, let's tackle them together. For bugs, head to [GitHub issues](https://github.com/teamhanko/hanko/issues/new/choose) and choose "Bug Report." Provide as much detail as you can; good bug reports are like gold!

### Security

For security concerns, shoot us an email at `security@hanko.io` instead of opening an issue. We'll handle it with the utmost priority.

### Bugs

Before reporting a bug:

1. Check [existing issues](https://github.com/teamhanko/hanko/issues).
2. Make sure it's not already fixed. Test with the latest `main` branch.
3. Ensure it's genuinely a bug. If unsure, reach out via [Security Section](https://github.com/teamhanko/hanko/security).

If you have solutions in mind, share them in the bug description.

## Feature Requests

Excited about a new feature? Awesome! Head to [GitHub issues](https://github.com/teamhanko/hanko/issues/new/choose), choose "Feature Request," and let us know your thoughts. Be detailed, and explain why it's a game-changer.

Before making a request:

1. Check [existing issues](https://github.com/teamhanko/hanko/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement).
2. Confirm it's not already implemented. Test with the latest `main` branch.

## Submitting Code

Ready to dive into code? Fantastic! Here's a quick guide:

1. [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) the [repository](https://github.com/teamhanko/hanko).
2. [Clone](https://docs.github.com/en/get-started/quickstart/fork-a-repo#cloning-your-forked-repository) your fork.
3. [Configure remotes](https://docs.github.com/en/get-started/quickstart/fork-a-repo#configuring-git-to-sync-your-fork-with-the-original-repository).
4. Create a new branch off `main`.

   ```bash
   git checkout -b <new-branch-name>
   ```

5. Make changes following our [Style Guidelines](#style-guidelines). Commit and push.
6. Update or add tests. Run locally first (`go test ./...` for `backend`).
7. If your change affects the `backend` API, update [Open API spec(s)](./docs/static/spec). For `backend` config changes, update [Config.md](./backend/docs/Config.md).
8. [Create a pull request](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests).
9. Fill out the [pull request template](./.github/PULL_REQUEST_TEMPLATE.md). If it's a work in progress, mark it as "Draft."

If tests fail, no worries! Address them until everything passes. Need help? Reach out on [Discussion Section](https://github.com/teamhanko/hanko/discussions).

## Commit Message Guidelines

Let's keep commit messages clear and structured:

```markdown
<type>(<optional scope>): <description>

<optional body>

<optional footer(s)>
```

Type options: build, ci, docs, feat, fix, perf, refactor, test, chore.

The headline should look like:

```markdown
<type>(<optional scope>): <description>
   │            │               │
   │            │               └─⫸ Summary in present tense. Not capitalized. No period at the end.
   │            │
   │            └─⫸ Commit Scope: optional
   │
   └─⫸ Commit Type: build|ci|docs|feat|fix|perf|refactor|test|chore
```

## Style Guidelines

For Go files, keep them formatted using [gofmt](https://go.dev/blog/gofmt).

```bash
# single file
go fmt path/to/changed/file.go

# all files, e.g., in 'backend' directory
go fmt ./...
```
