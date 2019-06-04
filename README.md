# GoAlert

GoAlert provides on-call scheduling, automated escalation, and notifications via SMS and voice to automatically engage the right person, the right way, and at the right time.

![image](https://user-images.githubusercontent.com/23565500/58896528-c1cdb100-86bb-11e9-96f1-a57198ece062.png)

## Installation

GoAlert is distributed as a single binary with release notes available from the [GitHub Releases](https://github.com/target/goalert/releases) page.

See our [Getting Started Guide](./docs/getting-started.md) for running GoAlert in a production environment.

### Quick Start

```bash
docker run -it --rm -p 8081:8081 goalert/all-in-one
```

GoAlert will be running at [127.0.0.1:8081](http://127.0.0.1:8081). You can login with `admin/admin123`.

## Contributing

If you'd like to contribute to GoAlert, please see our [Contributing Guidelines](./CONTRIBUTING.md) and the [Development Setup Guide](./docs/development-setup.md).

Please also see our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Contact Us

If you need help or have a question, visit the [#GoAlert](https://gophers.slack.com/messages/CJQGZPYLV/) channel on [Gophers Slack](https://gophers.slack.com/) (use the [invite app](https://invite.slack.golangbridge.org/) for access).

- Vote on existing [Feature Requests](https://github.com/target/goalert/issues?q=is%3Aopen+label%3Afeature-request+sort%3Areactions-%2B1-desc) or submit [a new one](https://github.com/target/goalert/issues/new)
- File a [bug report](https://github.com/target/goalert/issues)
- Report security issues to security@goalert.me

## License

GoAlert is licensed under the [Apache License, Version 2.0](./LICENSE.md).
