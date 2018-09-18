# Contributing Guidelines

## License and CLA

The Kubeless license is Apache Software License V2

We do not currently ask for a Contributor License Agreement to be signed.

## Support Channels

Whether you are a user or contributor, official support channels include:

- GitHub [issues](https://github.com/kubeless/kubeless/issues/new)
- Slack: #kubeless room in the [Kubernetes Slack](http://slack.k8s.io/)

Before opening a new issue or submitting a new pull request, it's helpful to search the project - it's likely that another user has already reported the issue you're facing, or it's a known issue that we're already aware of.

## How to become a contributor and submit your own code

### Setup your development environment

Consult the [Developer's guide](./docs/dev-guide.md) to setup yourself up.

### Contributing a patch

1. Submit an issue describing your proposed change to the repo in question.
2. The [repo owners](OWNERS) will respond to your issue promptly.
3. If your proposed change is accepted, fork the desired repo, develop and test your code changes.
4. Submit a pull request making sure you fill up clearly the description, point out the particular
   issue your PR is mitigating, and ask for code review. If the PR is related to Kafka, include at least the tag [Kafka] in the title. You will be asked to add tests (either unit or e2e tests depending on the patch) and update any affected documentation.

## Issues

Issues are used as the primary method for tracking anything to do with the Kubeless project.

### Issue Type

* Question: These are support or functionality inquiries that we want to have a record of for future reference. Generally these are questions that are too complex or large to store in the Slack channel or have particular interest to the community as a whole. Depending on the discussion, these can turn into "Feature" or "Bug" issues.
* Proposal: Used for items (like this one) that propose a new ideas or functionality that require a larger community discussion. This allows for feedback from others in the community before a feature is actually developed. This is not needed for small additions. Final word on whether or not a feature needs a proposal is up to the core maintainers. All issues that are proposals should both have a label and an issue title of "Proposal: [the rest of the title]." A proposal can become a "Feature" and does not require a milestone.
* Features: These track specific feature requests and ideas until they are complete. They can evolve from a "Proposal" or can be submitted individually depending on the size.
* Bugs: These track bugs with the code or problems with the documentation (i.e. missing or incomplete)


