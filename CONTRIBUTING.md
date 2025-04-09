# Contributing to TigrisFS

Thanks for your interest in contributing to **TigrisFS**!

## Setup

To set up the project locally:

```bash
make setup
```
This will install all the necessary dependencies and set repositories commit message hook.

## Running Tests
Run unit tests:

```bash
make run-lint
make run-test
make run-cluster-test
make run-xfstests
```

## We use Conventional Commits

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for commit messages.
This helps in generating a changelog and understanding the history of changes.

Examples:

  * feat(fs): add support for compressed file blocks
  * fix(cluster): handle node reconnection errors
  * chore: update dependencies

Common prefixes:

  * feat – new feature
  * fix – bug fix
  * docs – documentation only changes
  * style – formatting, missing semicolons, etc.
  * refactor – code change that neither fixes a bug nor adds a feature
  * perf – performance improvement
  * test – adding or fixing tests
  * chore – tooling or maintenance

## Pull Requests

Fork the repo and create your branch:

```bash
git checkout -b feat/my-feature
```

Make your changes and commit using Conventional Commits.

Run the tests to ensure nothing breaks.

Push to your fork and open a Pull Request.