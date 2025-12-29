# Infrastructure Layer

This package contains cross-cutting infrastructure concerns.

## Contents

- `logger/` - Structured logging with zerolog
- `timeutil/` - Clock interface and timezone caching
- `retry/` - Generic retry mechanism with exponential backoff

## Principles

- Utility functions should be stateless where possible
- Use interfaces for testability (e.g., Clock)
- Thread-safe implementations
