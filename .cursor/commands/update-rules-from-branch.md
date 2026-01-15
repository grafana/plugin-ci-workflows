---
description: Update cursor rules based on patterns introduced in the current branch
---

# Update Cursor Rules from Branch Changes

Compare the current git branch with the main branch and update the cursor rules files to reflect any new patterns, types, functions, or conventions introduced in the changes.

## Instructions

1. **Analyze the branch diff**: Run `git diff main...HEAD` to understand all changes made in this branch compared to main.

2. **Read existing cursor rules**: Read all rule files in:
   - `.cursor/rules/` (Cursor IDE rules with `.mdc` extension)
   - `.agent/rules/` (agent rules with `.md` extension, if present)

3. **Identify new patterns**: From the diff, extract:
   - New types, structs, or interfaces
   - New functions or methods (especially exported ones)
   - New testing patterns or helpers
   - New workflow options or mutators
   - New mock utilities
   - Changes to package organization
   - New conventions or best practices demonstrated in the code

4. **Update the appropriate rule files**:
   - For Go test framework changes (`tests/act/`): Update `go-tests.mdc` / `go-tests.md`
   - For GitHub Actions workflow changes (`.github/workflows/`, `actions/`): Update `github-actions.mdc` / `github-actions.md`
   - For project-wide changes: Update `project.mdc` / `project.md`

5. **Maintain consistency**: When updating both `.cursor/rules/*.mdc` and `.agent/rules/*.md` files:
   - Keep the content identical except for the frontmatter format
   - `.cursor/rules/*.mdc` uses: `title`, `description`, `globs`, `alwaysApply`
   - `.agent/rules/*.md` uses: `trigger: glob` or `trigger: always_on`, `globs`

6. **Add documentation for**:
   - New package files and their purpose
   - New exported types with their fields and usage examples
   - New functions with code examples showing how to use them
   - New options/patterns with before/after examples where helpful
   - Summary tables for groups of related options or helpers

7. **Do NOT**:
   - Remove existing documentation unless it's outdated
   - Add speculative documentation for code that doesn't exist yet
   - Document internal/unexported implementation details
   - Make assumptions about patterns not clearly demonstrated in the diff

## Output

After updating the rules, provide a summary of:
- Which rule files were updated
- What new sections or patterns were added
- Any significant changes to existing documentation
