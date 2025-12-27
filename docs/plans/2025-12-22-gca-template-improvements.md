# GCA Template Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add SKIP decision option and formatter instructions to GCA review template for consistency with CLI prompt.

**Architecture:** Update the `generatePatchPrompt` function in `internal/gca/fetch.go` to align the feedback template with the runtime instructions in `LaunchClaudeCode`. This ensures consistent guidance between the saved markdown files and the interactive Claude Code sessions.

**Tech Stack:** Go (template string generation)

---

## Task 1: Add SKIP Option to Decision List

**Files:**
- Modify: `internal/gca/fetch.go:285-287`

**Step 1: Add SKIP option to decision list**

Update lines 285-287 to include SKIP as a third decision option:

```go
2. **Decide** one of:
   - **APPLY:** Implement the suggestion (verbatim or with modifications)
   - **REJECT:** The feedback is incorrect, inapplicable, or low-value
   - **SKIP:** The target file doesn't exist or the feedback is no longer applicable
```

**Step 2: Update Act section to include SKIP**

Update lines 289-291 to handle all three decision types:

```go
3. **Act** on your decision:
   - If APPLY: Edit the file at `+"`%s`"+` and run relevant formatters (e.g., gofmt, prettier)
   - If REJECT or SKIP: Do not modify any files
```

**Step 3: Update decision format example**

Update line 296 to include SKIP in the format:

```go
## Decision: [APPLY | REJECT | SKIP]
```

**Step 4: Build and test the change**

```bash
make build
```

Expected: Build succeeds with no errors

**Step 5: Manually verify the template output**

Read the updated function to verify the changes:

```bash
grep -A 20 "2. \*\*Decide\*\*" internal/gca/fetch.go
```

Expected: Should show SKIP option in all three locations (decision list, act section, format example)

**Step 6: Commit the changes**

```bash
git add internal/gca/fetch.go
git commit -m "feat(gca): add SKIP decision option to review template

- Add SKIP as third decision type alongside APPLY/REJECT
- Update Act section to handle SKIP case
- Update decision format to include SKIP
- Add formatter mention (gofmt, prettier) in APPLY action
- Aligns template with LaunchClaudeCode prompt instructions"
```

---

## Task 2: Verify Consistency Between Template and CLI Prompt

**Files:**
- Read: `internal/gca/fetch.go:276-306`
- Read: `internal/gca/process.go:79-107`

**Step 1: Compare template and CLI prompt side-by-side**

```bash
# Extract template section
sed -n '276,306p' internal/gca/fetch.go > /tmp/template.txt

# Extract CLI prompt section
sed -n '79,107p' internal/gca/process.go > /tmp/cli_prompt.txt

# Review both
cat /tmp/template.txt
echo "---"
cat /tmp/cli_prompt.txt
```

Expected: Both should now mention:
- SKIP as a decision option
- Formatters/linters in the APPLY action
- File modification constraints

**Step 2: Test with mock data (optional manual test)**

If you want to verify the full workflow:

```bash
# This would require actual GitHub PR data
# Just verify the build works for now
./builds/smix --help
./builds/smix gca --help
```

Expected: Commands execute without errors

**Step 3: Clean up test artifacts**

```bash
rm -f /tmp/template.txt /tmp/cli_prompt.txt
```

**Step 4: Final verification**

```bash
go test ./internal/gca/...
```

Expected: All tests pass (if any exist)

**Step 5: Final commit (if any additional tweaks needed)**

```bash
git status
# If clean, no commit needed
# If additional changes, commit them
```

---

## Testing Notes

**Manual Testing:**
To fully test this change, you would need to:
1. Run `smix gca review owner/repo pr_number` against a real PR with gemini-code-assist feedback
2. Verify the generated markdown files include SKIP option
3. Verify the Claude Code session shows consistent instructions

**Automated Testing:**
Consider adding a unit test for `generatePatchPrompt` to verify output contains:
- "SKIP" option in decision list
- "formatter" mention in Act section
- "SKIP" in decision format example

---

## Completion Criteria

- [ ] Template includes SKIP as a decision option (line ~287)
- [ ] Act section mentions formatters and handles SKIP (line ~290)
- [ ] Decision format example includes SKIP (line ~296)
- [ ] Code builds successfully with `make build`
- [ ] Template and CLI prompt are consistent
- [ ] Changes are committed with descriptive message

---

## Notes

This is a text-only change to template generation - no logic changes, no API changes, no breaking changes. The risk is extremely low. The benefit is improved clarity and consistency for Claude Code sessions processing gemini-code-assist feedback.
