# Contributing

To contribute,
[fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo) this repository, commit your changes,
and [send a Pull Request](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests).

Note that our [code of conduct](https://github.com/newrelic/.github/blob/main/CODE_OF_CONDUCT.md)
applies to all platforms and venues related to this project; please follow it in all your
interactions with the project and its participants.

## Pull Requests

1. Ensure any install or build dependencies are removed before the end of the layer when doing a
   build.
2. Increase the version numbers in any examples files and the README.md to the new version that this
   Pull Request would represent. The versioning scheme we use is [SemVer](https://semver.org/).
3. You may merge the Pull Request in once you have the sign-off of two other developers, or if you
   do not have permission to do that, you may request the second reviewer to merge it for you.

### Commit Messages

Each commit message should follow
the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format. This format
provides a way to make the commit history more readable and easy to understand. It also allows for
automatic generation of changelogs.

The format of the commit header is as follows with the `<scope>` being optional:

```
<type>(<scope>): <subject>
```

#### Scope

Scope can be used to add context to the commit and is recommended when making changes that only
apply to specific distros.

#### Type

The type must be one of the following:

| Type         | Description                                                   | Changelog Required |
|--------------|---------------------------------------------------------------|--------------------|
| **build**    | Changes that affect the build system or external dependencies | no                 |
| **ci**       | Changes to CI configs and scripts                             | no                 |
| **docs**     | Documentation changes                                         | no                 |
| **feat**     | A new feature                                                 | yes                |
| **fix**      | A bug fix related to one of the distros                       | yes                |
| **perf**     | A performance enhancement                                     | yes                |
| **refactor** | A code change that neither fixes a bug nor adds a feature     | no                 |
| **style**    | Changes that effect only code style                           | no                 |
| **test**     | Adding or updating tests                                      | no                 |

### Changelog

Pull requests which include user-facing changes must be accompanied by a changelog entry. This repository uses OpenTelemetry's `chloggen` tool to manage changelog entries. To make a changelog entry:

1. Create an entry file using `make chlog-new`
2. Fill in the empty fields in the new file
3. Run `make chlog-validate` to validate the new file
4. Commit the file with your changes

If a changelog entry is not required, you may prefix your PR title with any of the conventional commit tags marked "no" in the table above, or add the `Skip Changelog` label to your pull request.

During the release process, all changelog entries are transcribed in CHANGELOG.md and added to [NRDOT's release notes](https://docs.newrelic.com/docs/release-notes/nrdot-release-notes/).

## Contributor License Agreement

Keep in mind that when you submit your Pull Request, you'll need to sign the CLA via the
click-through using CLA-Assistant. If you'd like to execute our corporate CLA, or if you have any
questions, please drop us an email at opensource@newrelic.com.

For more information about CLAs, please check out Alex Russell’s excellent post,
[“Why Do I Need to Sign This?”](https://infrequently.org/2008/06/why-do-i-need-to-sign-this/).
