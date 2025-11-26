# Session Notes: Secrets Detection Validator

Implementation details for Feature 4 of the improvements plan.

## Architecture

### Two-Tier Detection

1. **Fast Path**: Pre-compiled regex patterns (always runs, CPU-bound)
2. **Slow Path**: gitleaks binary (optional, when available and configured)

### Package Structure

```text
internal/validators/secrets/
├── patterns.go      # Pattern definitions with regex and metadata
├── detector.go      # Detector interface and PatternDetector implementation
├── validator.go     # SecretsValidator implementing validator.Validator
├── *_test.go        # Tests (62 tests, 92.6% coverage)

internal/linters/
└── gitleaks.go      # GitleaksChecker interface and implementation

pkg/config/
└── secrets.go       # SecretsConfig and SecretsValidatorConfig schemas

internal/config/factory/
└── secrets_factory.go  # SecretsValidatorFactory
```

## Pattern Definition

Patterns are defined with metadata for better error reporting:

```go
type Pattern struct {
    Name        string           // e.g., "aws-access-key-id"
    Description string           // e.g., "AWS Access Key ID"
    Regex       *regexp.Regexp   // Compiled pattern
    ErrorCode   validator.ErrorCode  // e.g., ErrSecretsAPIKey
}
```

## Built-in Patterns (25+)

- **AWS**: Access Key ID (`AKIA...`), Secret Access Key
- **GitHub**: PAT (`ghp_`), OAuth (`gho_`), App (`ghs_`/`ghu_`), Refresh (`ghr_`)
- **GitLab**: PAT (`glpat-`)
- **Slack**: Tokens (`xox[baprs]-`), Webhook URLs
- **Google/GCP**: API Keys (`AIza`), Service Account JSON
- **Private Keys**: RSA, DSA, EC, OpenSSH, PGP
- **Database**: MongoDB, PostgreSQL, MySQL, Redis connection strings
- **Generic**: Passwords, secrets, API keys (higher false positive risk)
- **Services**: NPM, Stripe, Twilio, SendGrid, Mailgun, Heroku, Azure
- **JWT**: JSON Web Tokens

## Configuration Schema

```toml
[validators.secrets.secrets]
enabled = true
use_gitleaks = false          # Enable gitleaks as second-tier
max_file_size = 1048576       # 1MB default, skip larger files
block_on_detection = true     # false = warn only

# Suppress false positives
allow_list = [
    "AKIA.*EXAMPLE",          # Test/example keys
    "ghp_test.*"              # Test tokens
]

# Disable specific patterns
disabled_patterns = [
    "generic-password",       # Too many false positives
]

# Add organization-specific patterns
[[validators.secrets.secrets.custom_patterns]]
name = "internal-api-key"
description = "Internal API Key"
regex = "MYCOMPANY_[A-Z0-9]{32}"
```

## Finding Representation

```go
type Finding struct {
    Pattern *Pattern  // Which pattern matched
    Match   string    // The actual matched text
    Line    int       // 1-indexed line number
    Column  int       // 1-indexed column number
}
```

## Factory Integration

The `SecretsValidatorFactory` creates validators with:

1. Pattern detector with default + custom patterns
2. Optional gitleaks checker
3. Predicate: `PreToolUse` + `Write|Edit` tools

```go
validator.And(
    validator.EventTypeIs(hook.EventTypePreToolUse),
    validator.ToolTypeIn(hook.ToolTypeWrite, hook.ToolTypeEdit),
)
```

## Content Extraction

- **Write**: Uses `hookCtx.GetContent()` directly
- **Edit**: Uses `hookCtx.ToolInput.NewString` (validates new content being written)

## Testing Patterns

When testing regex patterns, ensure test data matches exact pattern requirements:

```go
// Pattern: AIza[0-9A-Za-z_-]{35} requires exactly 35 chars after "AIza"
"AIzaSyD-abcdefghijklmnopqrstuvwxyz12345"  // 35 chars ✓

// Pattern: SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}
"SG.abcdefghijklmnopqrstuv.wxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abc"  // 22 + 43 chars ✓
```

## Error Codes

Uses existing SEC category codes from `error_code.go`:

- `SEC001` (ErrSecretsAPIKey) - API keys
- `SEC002` (ErrSecretsPassword) - Hardcoded passwords
- `SEC003` (ErrSecretsPrivKey) - Private keys
- `SEC004` (ErrSecretsToken) - Tokens (GitHub, Slack, JWT, etc.)
- `SEC005` (ErrSecretsConnString) - Database connection strings

## Parallel Execution

Returns `CategoryCPU` since regex matching is CPU-bound, not I/O-bound.

## Key Learnings

1. **ByteSize Type**: Created `config.ByteSize` for human-readable file sizes
2. **Logger Interface**: No `Warn` method exists - use `Info` or `Error`
3. **Interface Returns**: Use `//nolint:ireturn // interface for polymorphism` for factory methods
4. **Preallocation**: Always pre-allocate slices when length is known
5. **Unused Parameters**: Rename to `_` when parameter is required by interface but unused
