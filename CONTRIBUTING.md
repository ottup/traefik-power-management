# Contributing to Traefik Power Management Plugin

Thank you for your interest in contributing to the Traefik Power Management Plugin! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Contributions](#making-contributions)
- [Development Guidelines](#development-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow:

- **Be Respectful**: Treat all community members with respect and kindness
- **Be Collaborative**: Work together constructively and welcome diverse perspectives
- **Be Patient**: Remember that everyone has different experience levels and backgrounds
- **Be Constructive**: Provide helpful feedback and suggestions for improvement
- **Focus on the Project**: Keep discussions relevant to the project's goals and objectives

## Getting Started

### Prerequisites

Before you start contributing, make sure you have:

- **Go 1.21+**: Required for development and testing
- **Git**: For version control
- **Traefik**: For testing plugin integration (optional but recommended)
- **Basic Understanding**: Familiarity with Go, Traefik, and networking concepts

### Ways to Contribute

You can contribute to this project in several ways:

1. **Bug Reports**: Report bugs and issues you encounter
2. **Feature Requests**: Suggest new features and improvements
3. **Code Contributions**: Submit code changes via pull requests
4. **Documentation**: Improve documentation and examples
5. **Testing**: Help test new features and bug fixes
6. **Community Support**: Help other users in discussions and issues

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/your-username/traefik-power-management.git
cd traefik-power-management

# Add upstream remote
git remote add upstream https://github.com/ottup/traefik-power-management.git
```

### 2. Set Up Development Environment

```bash
# Install Go dependencies
go mod tidy

# Verify setup
go build .
go vet ./...
go fmt ./...
```

### 3. Set Up Testing Environment (Optional)

```bash
# Create test directory structure for local plugin testing
mkdir -p test/plugins-local/src/github.com/ottup/traefik-power-management

# Copy plugin files for testing
cp main.go go.mod .traefik.yml test/plugins-local/src/github.com/ottup/traefik-power-management/
```

### 4. Development Tools

Recommended tools for development:

```bash
# Install useful Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Making Contributions

### Creating Issues

When creating issues, please provide:

#### Bug Reports
- **Clear Title**: Concise description of the problem
- **Environment Details**: 
  - Traefik version
  - Plugin version
  - Operating system
  - Container platform (if applicable)
- **Configuration**: Sanitized configuration showing the issue
- **Steps to Reproduce**: Clear steps that reproduce the problem
- **Expected vs Actual Behavior**: What should happen vs what actually happens
- **Logs**: Relevant log output with debug mode enabled

#### Feature Requests
- **Clear Description**: What feature you'd like to see added
- **Use Case**: Why this feature would be valuable
- **Proposed Implementation**: Any ideas for how it could be implemented
- **Alternatives Considered**: Other approaches you've considered

### Working on Issues

1. **Check Existing Issues**: Look for existing issues before creating new ones
2. **Claim Issues**: Comment on issues you'd like to work on
3. **Ask Questions**: Don't hesitate to ask for clarification or guidance
4. **Start Small**: Begin with smaller issues to get familiar with the codebase

## Development Guidelines

### Code Standards

#### Go Code Style
Follow standard Go conventions:

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Run linter
golangci-lint run
```

#### Code Quality
- **Error Handling**: Implement comprehensive error handling
- **Logging**: Use appropriate log levels and structured logging
- **Comments**: Add meaningful comments for complex logic
- **Testing**: Include tests for new functionality

#### Plugin-Specific Guidelines
- **Yaegi Compatibility**: Ensure all code works in Traefik's Yaegi interpreter
- **Single File**: Keep plugin implementation in `main.go` for Yaegi compatibility
- **Standard Library**: Prefer Go standard library over external dependencies
- **Performance**: Optimize for minimal resource usage

### Configuration Changes

When modifying configuration:

1. **Backward Compatibility**: Maintain compatibility when possible
2. **Documentation**: Update configuration documentation
3. **Validation**: Add proper validation for new fields
4. **Examples**: Provide examples of new configuration options

### Security Considerations

- **Credential Handling**: Never log or expose credentials
- **Input Validation**: Validate all user inputs
- **Command Execution**: Safely handle command execution
- **Network Security**: Implement secure network communications

## Testing

### Local Testing

```bash
# Run Go tests (when available)
go test ./...

# Run static analysis
go vet ./...
staticcheck ./...

# Build and verify
go build .
```

### Integration Testing

For testing with Traefik:

```bash
# Set up local plugin testing
mkdir -p ./plugins-local/src/github.com/ottup/traefik-power-management
cp main.go go.mod .traefik.yml ./plugins-local/src/github.com/ottup/traefik-power-management/

# Create test Traefik configuration
cat > traefik-test.yml << EOF
experimental:
  localPlugins:
    traefik-power-management:
      moduleName: "github.com/ottup/traefik-power-management"

http:
  middlewares:
    test-wol:
      plugin:
        traefik-power-management:
          healthCheck: "http://test-target:3000/health"
          macAddress: "00:11:22:33:44:55"
          debug: true
EOF

# Run Traefik with test configuration
traefik --configFile=traefik-test.yml
```

### Testing Checklist

Before submitting changes:

- [ ] Code builds successfully (`go build .`)
- [ ] Code passes formatting (`go fmt ./...`)
- [ ] Code passes static analysis (`go vet ./...`)
- [ ] Plugin loads in Traefik without errors
- [ ] New features work as documented
- [ ] Existing functionality remains intact
- [ ] Configuration validation works properly
- [ ] Debug logging provides useful information

## Documentation

### Documentation Standards

- **Clear Language**: Use simple, clear language
- **Examples**: Provide working examples for all features
- **Structure**: Follow existing documentation structure
- **Links**: Use relative links for internal documentation

### Updating Documentation

When making changes that affect documentation:

1. **README**: Update main README if needed
2. **Configuration**: Update CONFIGURATION.md for config changes
3. **Deployment**: Update DEPLOYMENT.md for deployment-related changes
4. **Troubleshooting**: Add troubleshooting info for new features
5. **Changelog**: Add entries to CHANGELOG.md

### Documentation Files

- **README.md**: Main project overview and quick start
- **CONFIGURATION.md**: Comprehensive configuration reference
- **DEPLOYMENT.md**: Production deployment and security guide
- **TROUBLESHOOTING.md**: Common issues and solutions
- **CHANGELOG.md**: Version history and changes
- **CONTRIBUTING.md**: This file

## Pull Request Process

### Before Creating a Pull Request

1. **Sync with Upstream**: Make sure your fork is up to date
```bash
git checkout main
git pull upstream main
git push origin main
```

2. **Create Feature Branch**: Create a descriptive branch name
```bash
git checkout -b feature/your-feature-description
```

3. **Make Changes**: Implement your changes following the guidelines above

4. **Test Changes**: Ensure all tests pass and functionality works

### Pull Request Guidelines

#### PR Title and Description
- **Title**: Clear, concise description of changes
- **Description**: 
  - What changes were made and why
  - Any breaking changes
  - Testing performed
  - Related issues (use "Fixes #123" to auto-close issues)

#### PR Content
- **Small, Focused Changes**: Keep PRs focused on a single feature or fix
- **Clean Commits**: Use clear commit messages and squash if necessary
- **No Merge Commits**: Rebase instead of merging to keep history clean

#### Example PR Template
```markdown
## Description
Brief description of changes made.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that breaks existing functionality)
- [ ] Documentation update

## Testing Performed
- [ ] Local testing completed
- [ ] Integration testing with Traefik
- [ ] All existing functionality verified

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated if needed
- [ ] No breaking changes (or breaking changes documented)

## Related Issues
Fixes #123
```

### Review Process

1. **Automated Checks**: PRs must pass any automated checks
2. **Code Review**: Maintainers will review code and provide feedback
3. **Discussion**: Be responsive to feedback and questions
4. **Approval**: PRs need maintainer approval before merging
5. **Merge**: Maintainers will merge approved PRs

## Release Process

### Version Numbering

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes
- **MINOR** (0.X.0): New features (backward compatible)
- **PATCH** (0.0.X): Bug fixes (backward compatible)

### Release Workflow

1. **Version Bump**: Update version numbers in relevant files
2. **Changelog**: Update CHANGELOG.md with new version details
3. **Tag Release**: Create git tag with version number
4. **GitHub Release**: Create GitHub release with release notes
5. **Plugin Catalog**: Update Traefik plugin catalog (if applicable)

### Pre-release Testing

Before releases:
- Comprehensive testing with different Traefik versions
- Testing in various deployment scenarios (Docker, bare metal, etc.)
- Security review for any credential handling changes
- Documentation review and updates

## Getting Help

### Resources
- **Documentation**: Start with project documentation
- **Issues**: Search existing issues for similar problems
- **Discussions**: Use GitHub Discussions for questions and ideas

### Contact
- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: For security-related issues (contact maintainers privately)

### Community Guidelines
- Search before asking questions
- Provide context and details
- Be patient and respectful
- Help others when you can

## Recognition

Contributors who make significant contributions will be recognized:

- **Contributors File**: Added to CONTRIBUTORS.md (if created)
- **Release Notes**: Mentioned in release notes for significant contributions
- **GitHub**: GitHub's contribution tracking shows all contributions

## Thank You

Thank you for contributing to the Traefik Power Management Plugin! Your contributions help make this project better for everyone.

For questions about contributing, please create a [GitHub Discussion](https://github.com/ottup/traefik-power-management/discussions) or comment on relevant issues.