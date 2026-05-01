# OSS Launch Plan

This is the practical launch plan for making Faultline visible without damaging
developer trust. The goal is not empty traffic. The goal is credible technical
attention from Go, platform, DevEx, and architecture-minded engineers.

## Positioning

Primary message:

> Faultline finds structural risk in Go repositories locally, without uploading
> source code.

Use this framing consistently:

- Local-first structural risk analysis for Go codebases.
- Evidence-backed package risk reports.
- HTML, JSON, and SARIF output.
- No source upload, no runtime network dependency, no telemetry.
- Open-source scanner remains useful without login.

Avoid:

- Calling Faultline a vulnerability scanner.
- Promising that scores prove correctness or safety.
- Shaming public repositories.
- Asking for upvotes, stars, or coordinated voting.
- Implying cloud upload is required for useful results.

## Pre-Launch Checklist

- [ ] README first screen explains what Faultline does and how to run it.
- [ ] GitHub About description and topics are set.
- [ ] Issues and Discussions are enabled.
- [ ] Social preview asset is uploaded in repository settings.
- [ ] Example HTML and JSON reports are committed under `examples/reports`.
- [ ] At least one technical demo page exists under `docs/demo.md`.
- [ ] Release artifacts and Docker image are public and installable.
- [ ] Maintainer can respond to comments for a full launch day.

## Seven-Day Warmup

Ask 15-25 credible engineers for private feedback before posting broadly.

Good targets:

- Go maintainers
- platform engineers
- DevEx leads
- SREs
- monorepo owners
- people who write about technical debt, ownership, or architecture

Ask for feedback, not promotion:

```text
I’m building Faultline, a local-first Go repo risk scanner.
Would you run it on one repo and tell me whether the report is useful or noisy?

Command:
faultline scan ./... --format html --out faultline-report.html

No source code leaves your machine.
```

Success criteria before public launch:

- 5 concrete feedback items
- 3 issues or discussions from outside users
- 1 sample report that explains the product well
- 1 short list of known limitations in the launch post

## Launch Sequence

### Day 1: Hacker News

Use `Show HN` because Faultline is something people can try.

Suggested title:

```text
Show HN: Faultline – local structural risk reports for Go codebases
```

Use the first comment in [examples/launch/show-hn.md](../examples/launch/show-hn.md).

### Day 2-3: Go And Platform Communities

Use text-first posts. Do not cross-post the same link everywhere. Adapt the
message to each community and stay around to answer questions.

Good places to consider:

- Go community Slack or Discord
- `r/golang`, if the post is useful and follows community rules
- Platform engineering communities
- DevOps/SRE communities
- newsletters that cover Go, DevEx, CI, and code quality

### Day 4-5: Technical Thread

Publish a short technical thread built from screenshots and concrete examples:

1. The problem: repo-level quality signals hide package-level risk.
2. The method: churn, coverage, ownership, centrality, boundaries.
3. The trust model: local-first, no source upload.
4. The output: HTML, JSON, SARIF.
5. The ask: run it and report noisy findings.

### Day 7-10: Product Hunt

Use Product Hunt only after the README, demo, and early feedback loop are solid.
Product Hunt is useful for broader awareness, but the primary audience is still
technical buyers and practitioners.

Use [examples/launch/product-hunt.md](../examples/launch/product-hunt.md) as
the starting point.

## Follow-Up Content

Publish one short technical post every few days after launch:

- Why package-level risk beats repo-level quality scores.
- What CODEOWNERS misses in Go monorepos.
- Why churn plus low coverage matters more than complexity alone.
- How to fail CI only on new architecture violations.
- Local-first code analysis: what Faultline stores and what it never uploads.

Each post should include:

- one screenshot or report excerpt
- one command
- one limitation
- one concrete ask for feedback

## Success Metrics

Treat stars as awareness, not product validation.

Better first-month metrics:

- 15+ real issues or discussions
- 3+ outside pull requests
- 10+ teams trying SARIF or PR review
- 5+ design-partner conversations
- repeat scans from users who tried it once

## Maintainer Response Plan

On launch day:

- reply to every substantive question
- acknowledge noisy findings without defensiveness
- open issues from repeated feedback
- ship small docs fixes quickly
- avoid debating abstract scoring philosophy when a concrete fixture would help

The fastest way to earn trust is to be specific, transparent, and responsive.
