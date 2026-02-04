# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| < 0.2   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Atem, please report it responsibly.

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Email the maintainers directly at [your-email@example.com]
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Assessment**: We will assess the vulnerability and determine severity
- **Updates**: We will keep you informed of our progress
- **Resolution**: We aim to resolve critical issues within 7 days
- **Credit**: We will credit you in the release notes (unless you prefer anonymity)

## Security Considerations

### Workflow Execution

Atem executes workflows that may:

- Make HTTP requests to external services
- Execute shell commands
- Read and write files
- Connect to databases

**Recommendations:**

1. **Review workflows before running** - Especially workflows from untrusted sources
2. **Use profiles for secrets** - Never hardcode credentials in workflow files
3. **Limit file access** - Be cautious with file operations in workflows
4. **Validate inputs** - Sanitize user inputs in workflows

### Configuration Security

- Store sensitive data (API keys, database passwords) in profiles or environment variables
- Never commit `.env` files or profiles with secrets to version control
- Use environment variable references in profiles:

```yaml
secrets:
  api_key:
    env: MY_API_KEY
```

### Shell Node Security

The `shell` node executes arbitrary commands. Be aware that:

- Commands run with the same permissions as the Atem process
- Template variables in shell commands could be exploited if inputs aren't sanitized

**Safe usage:**
```yaml
# Use inputs carefully
- type: shell
  config:
    command: "echo '{{inputs.message}}'"  # Be cautious with user input
```

### Database Node Security

- Use parameterized queries (which Atem supports) to prevent SQL injection
- Limit database user permissions to only what's needed
- Use read-only connections where possible

```yaml
# Safe: parameterized query
- type: db
  config:
    sql: "SELECT * FROM users WHERE id = $1"
    params:
      - "{{inputs.user_id}}"

# Unsafe: string interpolation in SQL
- type: db
  config:
    sql: "SELECT * FROM users WHERE id = '{{inputs.user_id}}'"  # Don't do this!
```

## Best Practices

1. **Principle of least privilege** - Only grant necessary permissions
2. **Audit workflows** - Review what workflows do before running
3. **Secure secrets** - Use environment variables or secret managers
4. **Update regularly** - Keep Atem updated to the latest version
5. **Monitor execution** - Watch for unexpected behavior in workflows

## Known Limitations

- Atem does not sandbox workflow execution
- Shell commands have full user permissions
- No built-in secret encryption (use external tools)

## Security Updates

Security updates will be released as patch versions (e.g., 0.2.1) and announced in:

- GitHub Releases
- CHANGELOG.md
