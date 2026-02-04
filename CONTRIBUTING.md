# Contributing to Atem

Thank you for your interest in contributing to Atem! This document provides guidelines and information for contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/atem.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make test`
6. Run linter: `make lint`
7. Commit your changes (see commit guidelines below)
8. Push to your fork and submit a pull request

## Development Setup

### Prerequisites

- Go 1.24 or later
- golangci-lint (optional, for linting)

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run linter
make lint

# Run all checks
make check
```

## Project Structure

```
atem/
├── cmd/atem/     # Entry point
├── internal/        # Private packages
│   ├── config/      # Configuration loading
│   ├── engine/      # Workflow execution
│   └── tui/         # Terminal UI
├── nodes/           # Node implementations
├── pkg/node/        # Public Node interface
├── examples/        # Example workflows
└── docs/            # Documentation
```

## Adding a New Node Type

1. Create a new file in `nodes/` (e.g., `nodes/yournode.go`)
2. Implement the `pkg/node/Node` interface:

```go
package nodes

import "github.com/nameer-kp/atem/pkg/node"

type YourNode struct{}

func NewYourNode() *YourNode {
    return &YourNode{}
}

func (n *YourNode) Type() string {
    return "yourtype"
}

func (n *YourNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
    // Your implementation here
    return node.Result{
        Data: map[string]interface{}{
            "result": "value",
        },
    }, nil
}
```

3. Register it in `cmd/atem/main.go`:

```go
registry.Register(nodes.NewYourNode())
```

4. Add tests in `nodes/yournode_test.go`
5. Document the node in `docs/WIKI.md`

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(nodes): add redis node type
fix(engine): resolve template variables in nested objects
docs: update installation instructions
test(config): add loader tests for profile merging
```

## Pull Request Guidelines

1. **Title**: Use conventional commit format
2. **Description**: Explain what changes you made and why
3. **Tests**: Add tests for new functionality
4. **Documentation**: Update docs if needed
5. **Single focus**: Keep PRs focused on one thing

### PR Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated (if applicable)
- [ ] CHANGELOG.md updated (for user-facing changes)
- [ ] Commit messages follow conventions

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for non-obvious logic
- Use meaningful variable names

## Testing

- Write tests for new features
- Place tests in `*_test.go` files alongside the code
- Use table-driven tests where appropriate
- Mock external dependencies

```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"basic case", "input", "expected"},
        {"edge case", "", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Reporting Issues

When reporting issues, please include:

1. Atem version (`atem --version`)
2. Operating system and version
3. Steps to reproduce
4. Expected behavior
5. Actual behavior
6. Relevant workflow YAML (if applicable)

## Feature Requests

For feature requests:

1. Check existing issues first
2. Describe the use case
3. Explain why existing features don't solve it
4. Propose a solution (if you have one)

## Questions?

- Open a GitHub issue for questions
- Check existing issues and documentation first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
