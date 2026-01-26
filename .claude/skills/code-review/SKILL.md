---
name: code-review
description: Conduct a code review of the changes in the current branch compared to main
---

# Code Review

## Overview

Perform a thorough code review that verifies functionality, maintainability, and
security before approving a change. Focus on architecture, readability,
performance implications, and provide actionable suggestions for improvement.

Compare the current git branch with the main branch.

Do not run lint, vet, build or tests as part of the code review. Simply review the
changes, and consult other files in the repository as needed.

## Steps

1. **Understand the change**
    - If a PR exists for the current branch, use `gh pr view` to read the description and linked issues for context
    - Otherwise, rely on commit messages (`git log main..HEAD`) and any context provided by the user
    - Identify the scope of files and features impacted
    - Note any assumptions or questions to clarify with the author
2. **Validate functionality**
    - Confirm the code delivers the intended behavior
    - Exercise edge cases by guarding conditions mentally
    - Check error handling paths and logging for clarity
3. **Assess quality**
    - Ensure functions are focused, names are descriptive, and code is readable
    - Watch for duplication, dead code, or missing tests
    - Verify documentation and comments reflect the latest changes
4. **Review security and risk**
    - Look for injection points, insecure defaults, or missing validation
    - Confirm secrets or credentials are not exposed
    - Evaluate performance or scalability impacts of the change

## Review Checklist

### Functionality

- [ ] Intended behavior works and matches requirements
- [ ] Edge cases handled gracefully
- [ ] Error handling is appropriate and informative

### Code Quality

- [ ] Code structure is clear and maintainable
- [ ] No unnecessary duplication or dead code
- [ ] Tests/documentation updated as needed

### Security & Safety

- [ ] No obvious security vulnerabilities introduced
- [ ] Inputs validated and outputs sanitized
- [ ] Sensitive data handled correctly

### GitHub Actions (for workflow and action YAML files only)

These checks apply only to `.yml`/`.yaml` files in `.github/workflows/` and `actions/` directories:

- [ ] External actions pinned to full commit SHAs with version comment (e.g., `uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v4.2.2`)
- [ ] Same-repo references use `@main` (release process handles tagging)
- [ ] No GitHub expressions interpolated directly in `run:` blocks (use `env:` vars instead to prevent shell injection)
- [ ] Self-hosted runner labels used instead of `ubuntu-latest` (e.g., `ubuntu-arm64-small`)
- [ ] Complex logic uses `actions/github-script` rather than complex bash
- [ ] Inputs passed to shell via `env:` block, not inline `${{ }}`

## Additional Review Notes

- Architecture and design decisions considered
- Performance bottlenecks or regressions assessed
- Coding standards and best practices followed
- Resource management, error handling, and logging reviewed
- Suggested alternatives, additional test cases, or documentation updates
  captured

Provide constructive feedback with concrete examples and actionable guidance for
the author.