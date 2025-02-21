# Contributing

Contributions are always welcome. Before contributing please read the
[code of conduct](https://github.com/newrelic/.github/blob/main/CODE_OF_CONDUCT.md)
and [search the issue tracker](issues); your issue may have already been discussed or fixed in
`main`. To contribute,
[fork](https://help.github.com/articles/fork-a-repo/) this repository, commit your changes,
and [send a Pull Request](https://help.github.com/articles/using-pull-requests/).

Note that our [code of conduct](https://github.com/newrelic/.github/blob/main/CODE_OF_CONDUCT.md)
applies to all platforms and venues related to this project; please follow it in all your
interactions with the project and its participants.

## Feature Requests

Feature requests should be submitted in the [Issue tracker](../../issues), with a description of the
expected behavior & use case, where they’ll remain closed until sufficient
interest, [e.g. :+1: reactions](https://help.github.com/articles/about-discussions-in-issues-and-pull-requests/),
has
been [shown by the community](../../issues?q=label%3A%22votes+needed%22+sort%3Areactions-%2B1-desc).
Before submitting an Issue, please search for similar ones in the
[closed issues](../../issues?q=is%3Aissue+is%3Aclosed+label%3Aenhancement).

## Pull Requests

1. Ensure any install or build dependencies are removed before the end of the layer when doing a
   build.
2. Increase the version numbers in any examples files and the README.md to the new version that this
   Pull Request would represent. The versioning scheme we use is [SemVer](http://semver.org/).
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

| Type         | Description                                                   |
|--------------|---------------------------------------------------------------|
| **build**    | Changes that affect the build system or external dependencies |
| **ci**       | Changes to CI configs and scripts                             |
| **docs**     | Documentation changes                                         |
| **feat**     | A new feature                                                 |
| **fix**      | A bug fix related to one of the distros                       |
| **perf**     | A performance enhancement                                     |
| **refactor** | A code change that neither fixes a bug nor adds a feature     |
| **style**    | Changes that effect only code style                           |
| **test**     | Adding or updating tests                                      |

## Contributor License Agreement

Keep in mind that when you submit your Pull Request, you'll need to sign the CLA via the
click-through using CLA-Assistant. If you'd like to execute our corporate CLA, or if you have any
questions, please drop us an email at opensource@newrelic.com.

For more information about CLAs, please check out Alex Russell’s excellent post,
[“Why Do I Need to Sign This?”](https://infrequently.org/2008/06/why-do-i-need-to-sign-this/).

## Slack

We host a public Slack with a dedicated channel for contributors and maintainers of open source
projects hosted by New Relic. If you are contributing to this project, you're welcome to request
access to the #oss-contributors channel in the newrelicusers.slack.com workspace. To request access,
please use
this [link](https://join.slack.com/t/newrelicusers/shared_invite/zt-1ayj69rzm-~go~Eo1whIQGYnu3qi15ng).
