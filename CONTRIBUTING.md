# Contributing Guide

* [Ways to Contribute](#ways-to-contribute)
* [Development Environment Setup](#development-environment-setup)
* [Pull Request Lifecycle](#pull-request-lifecycle)
* [AI-Assisted Contributions](#ai-assisted-contributions)
* [Sign Your Commits](#sign-your-commits)
* [Ask for Help](#ask-for-help)

Welcome! We are so excited that you want to contribute to our project! 💖

As you get started, you are in the best position to give us feedback on areas of our project that we need help with including:

* Problems found during setting up a new developer environment
* Gaps in our guides or documentation
* Bugs in our tools and automation scripts

If anything doesn't make sense, or doesn't work when you try it, please open a bug report and let us know!

## Ways to Contribute

We welcome many different types of contributions including:

* [New features](https://github.com/kitops-ml/kitops/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* [Bug fixes](https://github.com/kitops-ml/kitops/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* [Documentation](https://github.com/kitops-ml/kitops/issues?q=is%3Aopen+is%3Aissue+label%3Adocumentation)
* [Builds and CI/CD](https://github.com/kitops-ml/kitops/issues?q=is%3Aopen+is%3Aissue+label%3Abuild)
* Release management
* Issue triage
* Answering questions on Discord, or the mailing list
* Web design
* Communications, social media, blog posts, or other marketing

If you think there's something else you can help with please contact us in the [#general channel of our Discord server](https://discord.gg/Tapeh8agYy) or during our [office hours meeting](https://github.com/kitops-ml/kitops/blob/main/GOVERNANCE.md#-meetings) and let's discuss how we can work together.

## Development Environment Setup

### Prerequisites

* Go (Golang): The latest version of Go, as the project is written in this language. Go's installation guide can be found at https://golang.org/doc/install.
* Git: Version control system for cloning the repository and managing code changes. Installation instructions are available at https://git-scm.com/book/en/v2/Getting-Started-Installing-Git.

### Setting up the project

1. Clone the Repository: Clone the KitOps source code to your local machine:

    ```shell
    git clone https://github.com/kitops-ml/kitops.git
    cd kitops
    ```

1. Install Go Dependencies: Inside the project directory, fetch and install the project's dependencies using the Go command:

    ```shell
    go mod tidy
    ```

1. Generate the dev mode harness and ui

    ```shell
    go generate ./...
    ```

1. Build the Kit CLI: Compile the source code into an executable named kit:

    ```shell
    go build -o kit
    ```

1. Run Your Build: Execute the built CLI to see all available commands:

    ```shell
    go run .
    ```

1. Updating Dependencies: If you add or update dependencies, ensure to update the go.mod and go.sum files by running `go mod tidy` again and include these changes in your commits.

## Pull Request Lifecycle

Pull requests are often called a "PR". KitOps generally follows the standard [GitHub pull request process](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests).

## AI-Assisted Contributions

We welcome the use of AI coding assistants (GitHub Copilot, Claude, ChatGPT, etc.) to help write code, documentation, and tests. However, you are fully responsible for reviewing and understanding all AI-generated code before submitting it.

### Requirements for AI-Assisted PRs

* **Review everything**: Read every line of AI-generated code. Understand what it does and why.
* **Test thoroughly**: AI can generate plausible-looking code that doesn't work or has subtle bugs. Run tests and verify behavior.
* **Remove AI artifacts**: Delete verbose comments, unnecessary explanations, or boilerplate that AI tends to add.
* **Check for hallucinations**: AI models can invent APIs, packages, or patterns that don't exist. Verify all imports and function calls.
* **Attest in your PR**: Use the AI-Assisted Code checklist in the PR template to acknowledge your review.

### What Reviewers Look For

Maintainers will be alert to common AI-generated code patterns:

* Overly defensive error handling or excessive nil checks
* Verbose or redundant comments explaining obvious operations
* Cargo-culted patterns that don't fit our codebase conventions
* Inconsistent style mixing modern and legacy approaches
* Test cases that pass but don't actually validate behavior

If a PR shows signs of unreviewed AI output, we'll reject it with feedback. This isn't punitive. It's about maintaining code quality and helping you learn to use AI tools effectively.

### Small PRs Principle

Keep PRs focused and reasonably sized. Smaller changes are easier to review (for both you and maintainers) and make it obvious when AI has generated too much code. If your AI assistant wants to refactor multiple files, split it into separate PRs.

## Code Reviews

There are two aspects of code review: giving and receiving. To make it easier for your PR to receive reviews, consider that reviewers will need you to:

* Follow the project coding conventions
* Write good commit messages
* Break large changes into a logical series of smaller patches which individually make easily understandable changes, and in aggregate solve a broader issue

Reviewers, the people giving the review, are highly encouraged to revisit our [Code of Conduct](https://github.com/kitops-ml/kitops/blob/main/CODE-OF-CONDUCT.md) and must go above and beyond to promote a collaborative, respectful community. When reviewing PRs from others [The Gentle Art of Patch Review](https://sage.thesharps.us/2014/09/01/the-gentle-art-of-patch-review/) suggests an iterative series of focuses which is designed to lead new contributors to positive collaboration without inundating them initially with a pile of suggestions:

1. Is the idea behind the contribution sound?
1. Is the contribution architected correctly?
1. Is the contribution polished?

If your PR isn't getting any attention after 3-4 days (remember Maintainers tend to be very busy) please ping one of the Maintainers in Discord.

## Sign Your Commits

Licensing is important to open source projects. It provides some assurances that the software will continue to be available based under the terms that the author(s) desired. We require that contributors sign off on commits submitted to our project's repositories. The [Developer Certificate of Origin (DCO)](https://probot.github.io/apps/dco/) is a way to certify that you wrote and have the right to contribute the code you are submitting to the project.

You sign-off by adding the following to your commit messages. Your sign-off must match the git user and email associated with the commit. Your commit message should be followed by:

    Signed-off-by: Your Name <your.name@example.com>

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running

    git commit --amend -s

## Ask for Help

The best way to reach us with a question when contributing is to ask on:

* The original github issue
* Our [Discord server](https://discord.gg/Tapeh8agYy)
* At our [office hours meeting](https://github.com/kitops-ml/kitops/blob/main/GOVERNANCE.md#-meetings)
