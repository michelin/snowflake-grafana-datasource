# Contributing Guidelines

_Pull requests, bug reports, and all other forms of contribution are welcomed and highly encouraged!_

## Contents

- [Contributing Guidelines](#contributing-guidelines)
    - [Contents](#contents)
    - [Code of Conduct](#code-of-conduct)
    - [Asking Questions](#asking-questions)
    - [Opening an Issue](#opening-an-issue)
        - [Reporting Security Issues](#reporting-security-issues)
        - [Bug Reports and Other Issues](#bug-reports-and-other-issues)
    - [Feature Requests](#feature-requests)
    - [Triaging Issues](#triaging-issues)
    - [Submitting Pull Requests](#submitting-pull-requests)
    - [Writing Commit Messages](#writing-commit-messages)
    - [Code Review](#code-review)
    - [Coding Style](#coding-style)
    - [Certificate of Origin](#certificate-of-origin)
    - [Credits](#credits)

> **This guide serves to set clear expectations for everyone involved with the project so that we can improve it together while also creating a welcoming space for everyone to participate. Following these guidelines will help ensure a positive experience for contributors and maintainers.**

## Code of Conduct

Please review our [Code of Conduct](CODE_OF_CONDUCT.md). It is in effect at all times. We expect it to be honored by everyone who contributes to this project. Inappropriate behavior will not be tolerated.

## Asking Questions

We don't use GitHub as a support forum. In short, GitHub issues are not the appropriate place to debug your specific project, but should be reserved for filing bugs and feature requests.

## Opening an Issue

Before [creating an issue](https://help.github.com/en/github/managing-your-work-on-github/creating-an-issue), check if you are using the latest version of the project. If you are not up-to-date, see if updating fixes your issue first.

### Reporting Security Issues

**Do not** file a public issue for security vulnerabilities. Instead, send your report privately to `TODO: Insert CERT email address`. This will help ensure that any vulnerabilities that are found can be disclosed responsibly to any affected parties.
We take [responsible disclosure](https://en.wikipedia.org/wiki/Responsible_disclosure) seriously, as they can affect confidence/trust in our project. The person who notified us of the vulnerability will be informed promptly that the problem is being addressed. Wherever possible, and depending on its criticality, the vulnerability should be corrected within the allotted time (generally between 90 and 120 days).

### Bug Reports and Other Issues

A great way to contribute to the project is to send a detailed issue when you encounter a problem. We always appreciate a well-written, thorough bug report. :v:

In short, since you are most likely a developer, **provide a ticket that you would like to receive**.

- **Review the documentation** before opening a new issue.
- **Do not open a duplicate issue**. Search through existing issues to see if your issue has previously been reported. If your issue exists, comment with any additional information you have. You may simply note "_I have this problem too_", which helps prioritize the most common problems and requests.
- **Prefer using [reactions](https://github.blog/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/)**, not comments, if you simply want to "+1" an existing issue.
- **Fully complete the provided issue template.** The bug report template requests all the information we need to quickly and efficiently address your issue. Be clear, concise, and descriptive. Provide as much information as you can, including steps to reproduce, stack traces, compiler errors, library versions, OS versions, and screenshots (if applicable).
- **Use [GitHub-flavored Markdown](https://help.github.com/en/github/writing-on-github/basic-writing-and-formatting-syntax).** Especially put code blocks and console outputs in backticks (```). This improves readability.

## Feature Requests

Feature requests are welcome! While we will consider all requests, we cannot guarantee your request will be accepted. We want to avoid [feature creep](https://en.wikipedia.org/wiki/Feature_creep). Your idea may be great, but also out-of-scope for the project. If accepted, we cannot make any commitments regarding the timeline for implementation and release. However, you are welcome to submit a pull request to help!

- **Do not open a duplicate feature request**. Search for existing feature requests first. If you find your feature (or one very similar) previously requested, comment on that issue.
- **Fully complete the provided issue template**. The feature request template asks for all necessary information for us to begin a productive conversation.
- Be precise about the proposed outcome of the feature and how it relates to existing features. Include implementation details if possible.

## Triaging Issues

You can triage issues which may include reproducing bug reports or asking for additional information, such as version numbers or reproduction instructions. Any help you can provide to quickly resolve an issue is very much appreciated!

## Submitting Pull Requests

We **love** pull requests! Before [forking the repository](https://help.github.com/en/github/getting-started-with-github/fork-a-repo) and [creating a pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/proposing-changes-to-your-work-with-pull-requests) for non-trivial changes, it is usually best to first open an issue to discuss the changes, or discuss your intended approach for solving the problem in the comments for an existing issue.

The more you contribute the more responsibility you will earn. Leadership roles in this project are merit-based and earned by peer acclaim.

_Note: All contributions will be licensed under the project's license._

- **Smaller is better.** Submit **one** pull request per bug fix or feature. A pull request should contain isolated changes pertaining to a single bug fix or feature implementation. **Do not** refactor or reformat code that is unrelated to your change. It is better to **submit many small pull requests** rather than a single large one. Enormous pull requests will take enormous amounts of time to review, or may be rejected altogether.
- **Coordinate bigger changes.** For large and non-trivial changes, open an issue to discuss a strategy with the maintainers. Otherwise, you risk doing a lot of work for nothing!
- **Prioritize understanding over cleverness.** Write code clearly and concisely. Remember that source code usually gets written once and read often. Ensure the code is clear to the reader. The purpose and logic should be obvious to a reasonably skilled developer, otherwise you should add a comment that explains it.
- **Include test coverage.** Add unit tests or UI tests when possible. Follow existing patterns for implementing tests.
- **Add documentation.** Document your changes with code comments or in existing guides.
- **Update the CHANGELOG** for all enhancements and bug fixes. Include the corresponding issue number if one exists, and your GitHub username. (example: "- Fixed crash in profile view. #123 @jessesquires")
- **Use the repo's default branch.** Branch from and [submit your pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork) to the repo's default branch. Usually this is `main`, but it could be `dev`, `develop`, or `master`.
- **[Resolve any merge conflicts](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/resolving-a-merge-conflict-on-github)** that occur.
- **Promptly address any CI failures**. If your pull request fails to build or pass tests, please push another commit to fix it.

## Writing Commit Messages

Please [write a great commit message](https://chris.beams.io/posts/git-commit/).

1. Separate subject from body with a blank line
1. Limit the subject line to 50 characters
1. Use the imperative mood in the subject line (example: "Fix networking issue")
1. Wrap the body at about 72 characters
1. Use the body to explain **why**, _not what and how_ (the code shows that!)
1. If applicable, prefix the title with the relevant component name. (examples: "[Docs] Fix typo", "[Profile] Fix missing avatar")

```text
[TAG] Short summary of changes in 50 chars or less

Add a more detailed explanation here, if necessary. Possibly give 
some background about the issue being fixed, etc. The body of the 
commit message can be several paragraphs. Further paragraphs come 
after blank lines and please do proper word-wrap.

Wrap it to about 72 characters or so. In some contexts, 
the first line is treated as the subject of the commit and the 
rest of the text as the body. The blank line separating the summary 
from the body is critical (unless you omit the body entirely); 
various tools like `log`, `shortlog` and `rebase` can get confused 
if you run the two together.

Explain the problem that this commit is solving. Focus on why you
are making this change as opposed to how or what. The code explains 
how or what. Reviewers and your future self can read the patch, 
but might not understand why a particular solution was implemented.
Are there side effects or other unintuitive consequences of this
change? Here's the place to explain them.

 - Bullet points are okay, too

 - A hyphen or asterisk should be used for the bullet, preceded
   by a single space, with blank lines in between

Note the fixed or relevant GitHub issues at the end:

Resolves: #123
See also: #456, #789
```

## Code Review

- **Review the code, not the author.** Look for and suggest improvements without disparaging or insulting the author. Provide actionable feedback and explain your reasoning.
- **You are not your code.** When your code is critiqued, questioned, or constructively criticized, remember that you are not your code. Do not take code review personally.
- **Always do your best.** No one writes bugs on purpose. Do your best, and learn from your mistakes.
- Kindly note any violations to the guidelines specified in this document.

## Coding Style

Consistency is the most important. Follow the[existing coding style, formatting, and naming conventions of the file you are modifying and of the overall project. When possible, these will be enforced with a linter. Failure to do so will result in a prolonged review process that has to focus on updating the superficial aspects of your code, rather than improving its functionality and performance.

Adhering to this code style ensures consistency and helps maintain code quality throughout the project.

## Certificate of Origin

From [_Developer's Certificate of Origin 1.1_](https://developercertificate.org/)

By making a contribution to this project, I certify that:

> 1. The contribution was created in whole or in part by me and I have the right to submit it under the open source license indicated in the file; or
> 1. The contribution is based upon previous work that, to the best of my knowledge, is covered under an appropriate open source license and I have the right under that license to submit that work with modifications, whether created in whole or in part by me, under the same open source license (unless I am permitted to submit under a different license), as indicated in the file; or
> 1. The contribution was provided directly to me by some other person who certified (1), (2) or (3) and I have not modified it.
> 1. I understand and agree that this project and the contribution are public and that a record of the contribution (including all personal information I submit with it, including my sign-off) is maintained indefinitely and may be redistributed consistent with this project or the open source license(s) involved.

## Credits

Based on [work](https://github.com/jessesquires/.github/blob/main/CONTRIBUTING.md) by [@jessesquires](https://github.com/jessesquires).