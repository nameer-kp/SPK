---
name: raise-pr
description: Create a GitHub pull request for the current branch with a well-structured description
arguments:
  - name: title
    description: Optional PR title (auto-generated from commits if not provided)
    required: false
  - name: base
    description: Base branch to merge into (defaults to master)
    required: false
disable-model-invocation: true
---

# Raise Pull Request

Create a GitHub pull request for the current branch.

## Workflow

1. **Gather context** - Run these commands to understand the changes:
   ```bash
   git status
   git branch --show-current
   git log master..HEAD --oneline
   git diff master...HEAD --stat
   ```

2. **Check prerequisites**:
   - Ensure there are commits to include (branch differs from base)
   - Ensure branch is pushed to remote (push with `-u` if needed)
   - Ensure `gh` CLI is authenticated

3. **Generate PR content**:
   - **Title**: Use provided title, or summarize from commit messages (under 70 chars)
   - **Body**: Structure as below

4. **Create PR** using `gh pr create`

## PR Body Template

```markdown
## Summary
[1-3 bullet points describing what this PR does]

## Changes
[List of files/areas changed with brief descriptions]

## Test plan
- [ ] [How to test the changes]

---
🤖 Generated with [Claude Code](https://claude.ai/code)
```

## Command

```bash
gh pr create --title "{{title}}" --base "{{base | default: master}}" --body "$(cat <<'EOF'
## Summary
...

## Changes
...

## Test plan
- [ ] ...

---
🤖 Generated with [Claude Code](https://claude.ai/code)
EOF
)"
```

## After Creation

- Output the PR URL so user can review
- Mention if CI checks are configured
