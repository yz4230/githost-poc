AGENTS.md

Purpose
-------
This document explains how to interact with automated coding agents (e.g. Copilot-style coding assistants) when working in this repository. It covers what context to provide, how to request changes, safety and review practices, and includes request templates and examples.

When to call an agent
---------------------
- Implementing or fixing code across multiple files.
- Adding new features that require repository-wide changes (DI registration, wiring, migrations).
- Creating tests, build scripts, or CI configurations.
- Refactoring where correctness should be validated by build/tests.

What to provide in a request
----------------------------
Always include the following to get accurate results:
- A concise description of the goal or bug to fix.
- The exact file(s) to change, or note that the agent should find the right files.
- Any constraints (Go version, frameworks, libraries, coding style).
- Desired behavior, including inputs/outputs and error modes.
- Whether to run the build/tests and report results.

Architecture Overview
--------------------
This is a Go-based Git hosting server following Clean Architecture principles with dependency injection.

**Tech Stack:**
- Web Framework: Echo v4 (`github.com/labstack/echo/v4`)
- DI Container: `github.com/samber/do`
- Database: SQLite with GORM (`gorm.io/gorm`, `gorm.io/driver/sqlite`)
- Logging: zerolog (`github.com/rs/zerolog`)
- Go Version: 1.25.1

**Layer Structure:**
```
internal/
├── entity/           # Domain entities and business errors
├── usecase/          # Business logic (interface + implementation)
├── repository/       # Data access layer (interface + GORM implementation)
├── storage/          # Git bare repository storage
├── git/              # Git smart HTTP protocol handlers
├── server/           # HTTP server setup and DI configuration
│   └── routes/       # HTTP route handlers
└── utils/            # Shared utilities
```

**Data Flow for API Requests:**
1. HTTP Request → `routes/*.go` (Echo handlers)
2. Handler invokes Usecase via DI: `do.MustInvoke[usecase.XxxUsecase](injector)`
3. Usecase executes business logic, calls Repository
4. Repository performs database operations via GORM
5. Response flows back through layers

**Entity Definitions:**
Core entities are in `internal/entity/`. Example: `entity.Repository` has fields like `ID`, `Name`, `Description`, `DeployBranch`, `LatestSHA`, `CreatedAt`, `UpdatedAt`.

**Standard Error Types** (defined in `internal/entity/err.go`):
- `entity.ErrNotFound` - Resource not found (return HTTP 404)
- `entity.ErrInvalid` - Validation failed (return HTTP 400)
- `entity.ErrConflict` - Resource conflict, e.g., duplicate name (return HTTP 409)
- `entity.ErrForbidden` - Access denied (return HTTP 403)
- `entity.ErrInternal` - Internal server error (return HTTP 500)

Repository-specific conventions
-------------------------------
**Dependency Injection:**
- DI is provided by `github.com/samber/do`.
- All dependencies are registered in `internal/server/server.go` in the `injectDependencies()` method.
- New constructors must be registered via `do.Provide(injector, <Constructor>)`.
- Constructors have signature: `NewXxx(injector *do.Injector) (Interface, error)`.

**Usecase Pattern:**
- Usecases live in `internal/usecase/` and follow this pattern:
  1. Define an interface: `type XxxUsecase interface { Execute(ctx context.Context, ...) (..., error) }`
  2. Implement with struct: `type xxxUsecaseImpl struct { ... }`
  3. Provide constructor: `func NewXxxUsecase(injector *do.Injector) (XxxUsecase, error)`
- Usecase implementations receive dependencies through constructor via DI.
- Example: See `internal/usecase/list_repository_usecase.go`.

**Repository Pattern:**
- Repositories live in `internal/repository/` and define data access interfaces.
- Implementations use GORM with helper: `gorm.G[ModelType](r.db)`.
- Model structs (e.g., `Repository`) have `FromEntity()` and `ToEntity()` conversion methods.
- Map GORM errors to entity errors: `gorm.ErrRecordNotFound` → `entity.ErrNotFound`.
- Constructor signature: `NewXxxRepository(i *do.Injector) (XxxRepository, error)`.

**HTTP Routes:**
- Routes are registered in `internal/server/routes/*.go`.
- Routes are grouped: `/api` for REST APIs, git smart HTTP for Git protocol.
- Handlers retrieve usecases via `do.MustInvoke[usecase.XxxUsecase](injector)`.
- Return appropriate HTTP status codes based on entity errors (see error mapping above).
- JSON request/response with inline structs for simple DTOs.

**Naming Conventions:**
- Entities: `entity.Repository`, `entity.Deployment`
- Usecases: `CreateRepositoryUsecase`, `GetRepositoryUsecase`, etc.
- Repositories: `RepositoryRepository`, `DeploymentRepository`
- Constructors: `NewXxxUsecase`, `NewXxxRepository`
- Private implementations: `xxxUsecaseImpl`, `xxxRepositoryImpl`

Safety & Review
---------------
- Agents may make large-scale edits. Always review diffs before committing or merging.
- Run `go build ./...` and any relevant tests after changes. The agent should run the build and report results.
- Prefer small, incremental changes rather than large sweeping edits.

Request templates
-----------------
### 1) Implement a new usecase
"Implement `GetXUsecase` in `internal/usecase/get_x_usecase.go` following the pattern used by `ListRepositoryUsecase`. The usecase should accept a `name` parameter and return the entity or `entity.ErrNotFound`. Register it in DI at `internal/server/server.go` in the `injectDependencies()` method. Run `go build ./...` to verify."

### 2) Fix a failing build
"Build is failing with these errors: <paste errors>. Please fix the code so `go build ./...` succeeds, and explain the root cause and the changes made."

### 3) Add a complete REST API endpoint
"Add endpoint `POST /api/widgets` that accepts JSON `{name: string, size: int}` and creates a widget. 

Steps:
1. Define `entity.Widget` in `internal/entity/widget.go` with ID, Name, Size, CreatedAt, UpdatedAt fields.
2. Create GORM model in `internal/repository/models.go` with `FromEntity()` and `ToEntity()` methods.
3. Define `WidgetRepository` interface in `internal/repository/widget.go` with `Create()` method and implement it.
4. Create `CreateWidgetUsecase` in `internal/usecase/create_widget_usecase.go` that validates input and calls repository.
5. Register both repository and usecase constructors in `internal/server/server.go`.
6. Add POST route handler in `internal/server/routes/api.go` that invokes the usecase and maps errors to HTTP status codes.
7. Run `go build ./...` and report results."

### 4) Add a new GET endpoint to existing API
"Add endpoint `GET /api/repositories/:name/deployments` that returns all deployments for a repository.

Steps:
1. Add `ListByRepositoryID(ctx context.Context, repoID entity.ID) ([]*entity.Deployment, error)` method to `DeploymentRepository` interface and implementation in `internal/repository/deployment.go`.
2. Create `ListDeploymentsByRepositoryUsecase` in `internal/usecase/list_deployments_by_repository_usecase.go` that accepts repository name, fetches repository, then lists its deployments.
3. Register usecase in DI at `internal/server/server.go`.
4. Add GET route in `internal/server/routes/api.go` that returns JSON array of deployments with proper error handling.
5. Run `go build ./...` to verify."

### 5) Add database migration
"Add a new field `IsPrivate bool` to the `entity.Repository` and update all related code:

Steps:
1. Add field to `entity.Repository` struct in `internal/entity/repository.go`.
2. Add column to GORM model in `internal/repository/models.go`.
3. Update `FillDefaults()` if needed.
4. GORM will auto-migrate on server start.
5. Run `go build ./...` to verify."

Examples
--------
### Good request (detailed):
"Add endpoint `GET /api/repositories/:name/stats` that returns statistics for a repository.

Steps:
1. Define a `RepositoryStats` struct in `internal/entity/repository.go` with fields: CommitCount, BranchCount, Size (int64).
2. Add `GetStats(ctx context.Context, repoID entity.ID) (*RepositoryStats, error)` method to `storage.GitStorage` interface and implementation.
3. Create `GetRepositoryStatsUsecase` that accepts repository name, looks up the repository, calls storage to get stats, and returns them.
4. Register in DI and add GET route handler.
5. Run `go build ./...` and verify."

### Bad request (too vague):
"Fix the app" - Missing: what needs fixing? which files? what's the error? what's the desired behavior?

### Good request (bug fix):
"The `POST /api/repositories` endpoint returns 500 when name contains spaces. Expected behavior: return 400 with validation error. Files likely involved: `internal/usecase/create_repository_usecase.go`. Run `go build ./...` after fix and add test case."

### Complete API creation example:
"Create a tags API for repositories with the following endpoints:
- `GET /api/repositories/:name/tags` - list all tags
- `POST /api/repositories/:name/tags` - create a new tag
- `DELETE /api/repositories/:name/tags/:tag` - delete a tag

Requirements:
1. Define `entity.Tag` with ID, RepositoryID, Name, SHA, CreatedAt
2. Implement `TagRepository` with CRUD operations
3. Implement three usecases: ListTags, CreateTag, DeleteTag
4. All usecases should verify repository exists first
5. Add routes in `internal/server/routes/api.go`
6. Map errors: tag not found → 404, duplicate tag → 409
7. Register all components in DI
8. Run `go build ./...` and report results"

Agent interaction rules for maintainers
-------------------------------------
- The agent may run commands and create files; treat its output as a patch candidate. Always review and run the build locally before merging.
- If the agent modifies tests or introduces new dependencies, verify CI config and lockfiles.

Maintainer checklist after an agent run
--------------------------------------
- Review changed files and git diff.
- Run `go build ./...` locally.
- Run tests (if present): `go test ./...`.
- If all checks pass, open a PR with a clear description of the changes.

Contact
-------
For questions about how to use agents in this repo, contact the repository owner or the maintainer team.
