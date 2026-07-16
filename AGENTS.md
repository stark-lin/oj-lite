# Agent Instructions

## Documentation

- When a change affects documentation, update the relevant documentation in the same change. This includes changes to behavior, APIs, configuration, setup, workflows, examples, or other user-facing information.

## Code Style

- Match the existing codebase style first. Before making changes, inspect nearby files and analogous modules, then follow the established package structure, file responsibilities, naming, control flow, error handling, logging, SQL, and testing patterns.
- Keep changes consistent with the surrounding code and avoid unrelated refactors, new abstractions, or introducing a different style without a clear need.
- When the repository does not already establish a convention, use idiomatic Go and established Go best practices.
- Format all Go code with `gofmt`; use `goimports` when available to organize imports.
