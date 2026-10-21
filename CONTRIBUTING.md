# Contributing

Thanks for your interest in improving this Bitrise step!

## Getting started

1. Fork the repository and clone your fork.
2. Make your changes — the step's code lives in `main.go`, its configuration in `step.yml`.
3. Test your changes locally:

   ```
   bitrise run test
   ```

   *Check `bitrise.yml` for any inputs that must be set in `.bitrise.secrets.yml`.*

## Submitting changes

1. Create a branch for your change.
2. Keep commits focused and write a clear commit message.
3. Open a pull request against `master` describing what changed and why, and how you tested it.

For general guidance on developing Bitrise steps, see the [Step Development Guideline](https://github.com/bitrise-io/bitrise/blob/master/_docs/step-development-guideline.md).

## Reporting issues

If you found a bug or have a feature request, please open an issue with steps to reproduce it, including your input values and relevant logs.
