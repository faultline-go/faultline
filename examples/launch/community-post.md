# Community Post Draft

Use this as a starting point for Go, platform, SRE, or DevEx communities. Adapt
it to the community and avoid posting the same text everywhere.

```text
I’m looking for feedback on Faultline, an open-source local CLI for structural
risk analysis in Go repositories.

It scans packages and explains risk using code structure, git churn, coverage,
ownership, dependency centrality, architecture boundary rules, and Go module
dependency metadata. Outputs are HTML, JSON, and SARIF.

The important bit for me: it runs locally and does not upload source code.

Try:

go install github.com/faultline-go/faultline/cmd/faultline@latest
faultline scan ./... --format html --out faultline-report.html

I’m especially interested in whether the report is actionable or noisy on real
Go repos, especially monorepos or repos with CODEOWNERS.
```

Good follow-up questions:

- Which finding would you ignore first?
- Which package surprised you?
- Did the ownership evidence match how your team actually works?
- Would SARIF annotations or PR markdown be more useful in your workflow?
