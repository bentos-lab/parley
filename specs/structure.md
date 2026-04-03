# Project Structure

This project follows Clean Architecture with clear separation of responsibilities.

- `adapter`: Handles communication with external systems.
- `adapter/inbound`: HTTP, CLI, RPC, WebSocket, subscribers, and similar interfaces.
- `adapter/outbound`: Databases, filesystems, cloud services, external APIs, caches, and similar integrations.
- `build`: Build automation scripts (see `build/build.sh` and `specs/build.md`) for producing the embedded SPA assets and Go binaries.
- `cmd`: Application entry points. Bootstraps the app and starts servers or workers. No business logic is allowed.
- `config`: Configuration definitions and loading logic. No business logic is allowed.
- `core`: Core logic and models.
- `shared`: Reusable utilities and helpers not related to business logic.
- `wiring`: Dependency injection and binding interfaces to concrete implementations. No business logic is allowed.

## Request Flow

- `cmd` (entry point) -> inbound adapter -> `core` -> outbound adapter.

## DTO

- Each inbound adapter defines request and response payload structures.
- Each outbound adapter defines structures for communicating with its external service.

## Usecase Contracts

### LLM

Input fields:

- `SystemInstruction`: System prompt or high-level instruction.
- `Messages`: Role-based conversation content.
- `Model`: Model identifier.
- `Temperature`: Sampling temperature (fixed per usecase in code).
- `MaxTokens`: Maximum tokens to generate (fixed per usecase in code).

Message fields:

- `Role`: Message role, supported values are `user` and `assistant`.
- `Content`: Message body text.

Methods:

- `Generate`: Produces free-form text.
- `GenerateJSON`: Produces output that is guaranteed to be JSON using a provided schema.
- `GenerateStream`: Produces streaming output.

## Outbound Adapter

### LLM

- Only OpenAI-compatible providers are supported initially (including endpoints for other providers such as Anthropic and Gemini).
