# Contributing to newrelic-istio-adapter

Welcome!
We gladly accept contributions from the community.

If you wish to contribute code and you have not signed our [Individual Contributor License Agreement](INDIVIDUAL_CLA.md) or our [Corporate Contributor License Agreement](CORPORATE_CLA.md), please do so in order to contribute.

## How to contribute

We use GitHub Pull Requests to incorporate code changes from external
contributors. 
Typical contribution flow steps are:

*   Sign the [Individual Contributor License Agreement](INDIVIDUAL_CLA.md) or our [Corporate Contributor License Agreement](CORPORATE_CLA.md).
*   Fork the `newrelic-istio-adapter` repository.
*   Clone the forked repository locally and configure the upstream repository.
*   Open an Issue describing what you propose to do (unless the change is so
    trivial that an issue is not needed).
    Issues should be submitted with a clear description, steps to reproduce the issue, Istio version used, and Kubernetes version.
*   Once you know which steps to take in your intended contribution, make changes
    in a topic branch and commit.
    (Don't forget to add or modify tests!).
*   Consult the [style guide](STYLE.md) and ensure your changes adhere to the project style.
*   Fetch changes from upstream, rebase with master and resolve any merge
    conflicts so that your topic branch is up-to-date.
*   Build and test the project locally.
    (See [DEVELOPMENT.md](DEVELOPMENT.md) for more details.)
*   Push all commits to the topic branch in your forked repository.
*   Submit a Pull Request to merge topic branch commits to upstream master.
    Be sure to reference your Issue when creating the PR unless it is a trivial change.

If this process sounds unfamiliar, have a look at the excellent overview of [collaboration via Pull Requests on GitHub](https://help.github.com/categories/collaborating-with-issues-and-pull-requests/)
for more information.

## Adding dependencies

Please avoid adding new dependencies to the project. If a change requires a new dependency, we encourage you to choose dependencies licensed under BSD, MIT, or Apache licenses.

## Do you have questions or are you experiencing unexpected behaviors after modifying this Open Source Software? 

Please engage with the “Build on New Relic” space in the [Explorers Hub](https://discuss.newrelic.com/c/build-on-new-relic/Integrations), New Relic’s Forum.
Posts are publicly viewable by anyone, please do not include PII or sensitive information in your forum post. 

## Contributor License Agreement ("CLA")

We'd love to get your contributions to improve `newrelic-istio-adapter`!
Keep in mind when you submit your pull request, you'll need to sign the [CLA](INDIVIDUAL_CLA.md) via the click-through using CLA-Assistant.
You only have to sign the CLA one time per project.

To execute our corporate CLA, which is required if your contribution is on behalf of a company, or if you have any questions, please drop us an email at open-source@newrelic.com. 

## Filing Issues & Bug Reports

We use GitHub issues to track public issues and bugs.
Issues should be submitted with a clear description, steps to reproduce the issue, Istio version used, and Kubernetes version.
If possible, please provide a link to an example app or gist that reproduces the issue.
Be aware that GitHub issues are publicly viewable by anyone, so please do not include personal information in your GitHub issue or in any of your contributions, except as minimally necessary for the purpose of supporting your issue.
New Relic will process any personal data you submit through GitHub issues in compliance with the [New Relic Privacy Notice](https://newrelic.com/termsandconditions/privacy).   

## A note about vulnerabilities  

New Relic is committed to the privacy and security of our customers and their data.
We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.
If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

## License

By contributing to `newrelic-istio-adapter`, you agree that your contributions will be licensed under the LICENSE file in the root directory of this source tree.

