# Agent Guideline

## Technique choice
- Language: `Golang`
- Testing framework: `stretchr/testify`
- HTTP Router: `chi`

## Project Description
- See `specs/project.md`

## Project Structure & Module Organization

- See `specs/structure.md`.

## Documentation
- Put all documentation, design notes, and guidelines under the `specs` directory.
- Exception: keep `README.md` and `AGENTS.md` in the repository root.
- Always add new environments to `.env.example`.
- Add comments to all functions describing what the function does, its parameters, and its return values.
- Add comments to ambiguous magic numbers, variables, and constants.
- Add brief inline comments for complex or non-obvious logic.

## Development Process

- When committing, load all changes to generate the commit message.
- Commit messages must include a title and a detailed description of the changes. The description must not exceed 72 characters.

## Coding Style and Naming Conventions

- Always use English in code.
- Follow standard Go conventions and idiomatic Go style.
- Keep files under ~700 LOC as a guideline (not a strict limit). Refactor or split files when it improves clarity or testability.
- For enums, prefer `type XXXEnum string` and define enum values on that type.
- Function and method names should be neutral when possible if they are not tied to business logic. If a function can be reused and is not tied to business logic, place it in `shared`.
- Go-context must be used to pass global values. Always use a private struct as key, sau đó viết 2 hàm để set và get value đó trong context. Các hàm này luôn được đặt ở nơi chính của value đó.

## Testing Guidelines

- Test file names must follow the `*_test.go` convention.
- Test modules should be named `<module>_test`.
- Only test exported/public functions and methods; validate private logic through public APIs.
- Tests should match the structure and naming of the source files.
- Pure test additions or fixes generally do not require a changelog entry, unless they affect user-facing behavior or the user explicitly requests one.

## Planning New Features

- When a new feature can reuse an existing function, use it. Update that function to be compatible with the new code and refactor any affected older code.
- Always run the following commands at the end of the implementation plan:
    + go test
    + go vet
    + staticcheck
    + gofmt
    + go mod tidy
