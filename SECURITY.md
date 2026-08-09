# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.7.x   | :white_check_mark: |
| 0.6.x   | :white_check_mark: (security fixes only) |
| < 0.6   | :x:                |

The latest release on the [Releases page](https://github.com/bdk38/HellPot/releases)
is the recommended deployment target.

## Reporting a Vulnerability

We prioritize the security of HellPot Community. Vulnerabilities often stem
from upstream dependencies, tools, or packages. We are committed to resolving
issues effectively.

Please follow these steps before reporting a potential vulnerability:

1. Verify that the behavior you've observed isn't already documented as normal
   HellPot behavior (for example, infinite response bodies on trap paths).
2. Clearly identify which component is vulnerable.
3. Provide a detailed description of the issue, including logs and, if
   available, debug output. Include all steps necessary to reproduce the
   vulnerability. If you have a proposed fix, please open a pull request.
4. Check whether the vulnerability is already known upstream. If there is an
   existing fix or patch, include that information in your report.

Prefer GitHub's private vulnerability reporting for this repository when
available; otherwise open a minimal public issue without exploit detail and
request a private follow-up channel.

## Dependency maintenance notes

- Direct Go module and GitHub Actions updates are handled via Dependabot.
- `github.com/mitchellh/reflectwalk` and `github.com/mitchellh/copystructure`
  are archived upstream. HellPot uses maintained forks under
  [`bdk38/reflectwalk`](https://github.com/bdk38/reflectwalk) and
  [`bdk38/copystructure`](https://github.com/bdk38/copystructure) via `replace`
  directives in `go.mod`.
