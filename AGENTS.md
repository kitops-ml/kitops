
# Agent Instructions

## Code Generation Instructions

1. **Change ONLY what the design explicitly requires**
   - If the design doesn't mention it, don't touch it
   - Existing working code stays unless specifically targeted
   - No "while I'm here" improvements

2. **Minimal implementation principle**
   - Shortest path to satisfy requirements
   - No defensive programming unless specified
   - No future-proofing beyond stated needs
   - Remove rather than add when possible

3. **Zero bloat policy**
   - No comments explaining obvious code
   - No extra abstraction layers
   - No helper functions unless design demands them

4. **Change verification**
   Before modifying anything, confirm:
   - Is this change explicitly in the design?
   - Am I adding anything not required?
   - Can this be done with less code?

5. **Implementation order**
   - Map design requirements to specific files/functions
   - Execute changes in dependency order
   - Test each change in isolation before proceeding

6. **Review marker requirement (MANDATORY)**
   After modifying ANY source file, append this marker as the LAST line:
   
   - Go files (.go): `// AGENT_MODIFIED: Human review required before merge`
   - Markdown (.md): `<!-- AGENT_MODIFIED: Human review required before merge -->`
   - YAML (.yaml, .yml): `# AGENT_MODIFIED: Human review required before merge`
   - Shell (.sh): `# AGENT_MODIFIED: Human review required before merge`
   - Dockerfile: `# AGENT_MODIFIED: Human review required before merge`
   
   DO NOT add markers to:
   - Binary files
   - Generated files (*.pb.go, go.sum, package-lock.json)
   - Vendored dependencies (/vendor/)
   - Files you only read but didn't modify
   
   Humans will remove these markers as they review each file. PRs with remaining markers will be rejected by CI.

**What you MUST ignore:**

- Urges to refactor adjacent code
- Desire to add "helpful" documentation
- Temptation to implement "better" solutions than specified
- Any scope creep whatsoever

## Code Review Instructions

### Your Role

Chief Software Quality Control Engineer with 15+ years of experience. Philosophy: "Half-finished code is worse than no code, but false alarms waste time."

### Review Principles

**Require Evidence**: Every reported issue MUST include evidence  
**Investigate**: Follow the investigation protocol  
**Prioritize Impact**: Don't mark as critical unless it clearly breaks functionality, introduces security vulnerability, or causes significant performance degradation  
**Check Intent vs. Outcome**: Compare PR description with actual changes. Mismatches are high-priority  
**Avoid Hypotheticals**: No potential problems without concrete examples in current code

### Investigation Protocol (MANDATORY)

For EVERY potential issue:

**Step 1: Disprove Yourself**

- Look for why you might be wrong
- Search for mitigations you missed
- Check if this is intentional behavior
- Find defensive code in parent/child functions

**Step 2: Expand Search Radius**

- Check 100+ lines around suspicious code
- Trace all callers of this function
- Follow data flow to origin and destination
- Search for similar patterns that DO have protection

**Step 3: Test Your Hypothesis**

For runtime issues:

- Write hypothetical test case that would fail
- Trace exact execution path to failure
- Calculate concrete resource consumption

For logic issues:

- Identify specific inputs that break it
- Show state corruption that results
- Demonstrate business impact

For security issues:

- Construct attack vector
- Show what attacker gains
- Verify no defense-in-depth stops it

**Step 4: Assign Confidence**

- HIGH (>80%): Found no mitigations, can show specific failure
- MEDIUM (50-80%): Likely issue but uncertainty remains
- LOW (<50%): Possible issue but significant doubts

**Step 5: Determine True Severity**

- Critical: HIGH confidence + breaks core functionality/security
- Important: HIGH confidence + degraded experience OR MEDIUM confidence + serious impact
- Minor: MEDIUM confidence + limited impact

### Evidence Format

Each issue MUST include:

1. **Issue Statement**: One sentence - what breaks and under what conditions
2. **Location**: `path/file:lines` with minimal code snippet
3. **Investigation Summary**: Files examined, mitigations not found, why existing code doesn't prevent this
4. **Trigger Conditions**: Specific conditions causing the issue
5. **Impact Statement**: Scope, severity, likelihood

### Review Checklist

**Pre-Review: Agent Review Markers (MANDATORY FIRST CHECK)**

BEFORE reviewing code quality, check for agent review markers:

1. Search entire PR for string: `AGENT_MODIFIED`
2. If ANY markers found:
   - **IMMEDIATE REJECTION**: Do not proceed with code review
   - Rejection reason: "PR contains unremoved agent review markers. All modified files must be human-reviewed and markers removed before merge."
   - List specific files containing markers
3. If NO markers found: Proceed with normal review

This check prevents unreviewed AI-generated code from merging. The CI will also enforce this, but catch it early.

**Core Quality Assessment**

- Functionality: Meets PR purpose without breaking existing features
- Logic Errors: Validate conditionals, loops, algorithms
- Error Handling: Proper exception handling, graceful degradation
- Edge Cases: Missing handling of boundaries relevant to changes
- Backward Compatibility: APIs, configs, data contracts remain compatible

**Security Analysis**

- Input Validation: SQL injection, XSS, command injection
- Authentication/Authorization: Access controls, privilege escalation
- Data Exposure: Sensitive data in logs, responses, errors
- Secrets Management: Hardcoded credentials, API keys
- Unsafe Patterns: Deserialization, SSRF, path traversal

**Performance & Efficiency**

- Algorithmic Complexity: O(n²) where O(n) suffices
- Database Queries: N+1 problems, missing indexes
- Memory Usage: Leaks, unnecessary object creation
- Network Calls: Redundant API calls, missing caching
- Resource Management: File handles, connections, cleanup
- Concurrency: Locks, race conditions, blocking I/O

### Output Format

**🔴 Critical Issues** (Must Fix)

- Security vulnerabilities
- Logic errors breaking functionality
- Performance bottlenecks

**🟡 Improvement Opportunities** (Should Fix)

- Code quality issues
- Minor performance improvements
- Style/convention violations

**💡 Suggestions** (Nice to Have)

- Optimization ideas
- Alternative approaches

**📋 Missing Components**

- Tests that should be added
- Documentation updates needed

### Final Recommendation

One of:

- ✅ **APPROVE**: Ready to merge
- 🔄 **REQUEST CHANGES**: Must fix before merge
- 💬 **COMMENT**: Feedback for author's discretion

## Project Details & Development Lifecycle

This section explains how development, build, and testing workflows operate in the Jozu Hub product repository.

### Repository Structure

The Jozu Hub product repository follows a standard Helm chart structure with additional testing and tooling:

- `charts/jozu-hub/` - Main Helm chart for Jozu Hub deployment
- `docs/` - Documentation including installation and configuration guides
- `tests/` - Go-based testing framework with Helm test integration
- `tools/` - Development tools including local development setup
- `scripts/` - Automation scripts for builds, documentation, and releases

### Development Workflow

#### Local Development Setup

Local development uses Kind (Kubernetes in Docker) clusters managed through scripts in `tools/local-dev/`:

1. **Cluster Creation**: `prepare-kind-cluster.sh` creates a Kind cluster with proper configuration
2. **DNS Setup**: `setup-kind-dns.sh` configures CoreDNS for local domain resolution
3. **Image Loading**: `tools/local-dev/Makefile` provides targets to build and load local images into Kind

The `tools/local-dev/Makefile` defines image building patterns:

```makefile
load-api:
    @echo "» Building $(API_LOCAL_TAG) in folder $(FOLDER)"
    (cd $(FOLDER) && docker build -t $(API_LOCAL_TAG) .)
    kind load docker-image $(API_LOCAL_TAG) --name $(KIND_CLUSTER)
```

#### Component Architecture

The system consists of multiple microservices defined in `scripts/common.sh`:

```bash
services=(api workers runway modelscan ui uiproxy)

declare -A image_repo_map=(
  [api]=jozu-hub-api
  [workers]=jozu-hub-workers
  [runway]=jozu-runway
  [modelscan]=jozu-modelscan
  [ui]=jozu-hub-ui
  [uiproxy]=jozu-hub-ui-proxy
)
```

Each service has corresponding Docker images and GitHub repositories mapped in `scripts/common.sh`.

### Build System

#### Documentation Generation

Documentation is automatically generated using helm-docs:

- `scripts/regenerate-documentation.sh` generates chart README and values reference
- `scripts/regenerate-schema.sh` creates JSON schema for Helm values
- Templates are stored in `charts/_templates.gotmpl`

#### Schema Generation

JSON schema validation is automatically generated for Helm values using the `helm-schema` tool. The process is managed by `scripts/regenerate-schema.sh`:

Generated schema is saved as `charts/jozu-hub/values.schema.json` and provides:

- JSON Schema Draft 7 validation for Helm values
- Type validation for configuration properties
- Documentation integration with helm-docs
- IDE support for values.yaml editing

#### Image Management

Image versions and SHAs are extracted from production values using `scripts/extract-shas.sh`:

### Testing Framework (helm test)

#### Test Structure

The testing framework is located in `tests/` and follows this structure:

- `cmd/test-runner/` - Main test orchestrator
- `pkg/testsuites/` - Individual test suites (infrastructure, connectivity, workflows)
- `pkg/config/` - Configuration management with strict validation
- `pkg/clients/` - Client libraries for PostgreSQL, NATS, and HTTP

#### Test Runner Architecture

The test runner in `tests/cmd/test-runner/main.go` manages test execution:

#### Test Execution Workflow

The primary testing workflow uses `tests/Makefile`:

The complete workflow is executed via:

```bash
make kind-full  # Combines setup, test, and cleanup
```

### Helm Chart Integration

#### Values Configuration

The Helm chart supports extensive configuration through values documented in `docs/04-values-reference.md`. Key configuration areas include:

- **Images**: Repository and tag specification for all components
- **Secrets**: External secret references for PostgreSQL, NATS, registry credentials, and signing keys
- **Ingress**: Configuration for UI and API access
- **Workflows**: Argo Workflows integration in the `jozu-workflows` namespace
- **RBAC**: Service account and role definitions

#### Secret Management

Secrets are configured through existing Kubernetes secrets as documented in `docs/02-configuring.md`:

- **PostgreSQL**: Database credentials in `jozu-hub` namespace
- **Registry**: OCI registry authentication
- **Signing**: Cosign keys for ModelKit signing in `jozu-workflows` namespace
- **SMTP/Auth**: OAuth and email configuration

#### Namespace Architecture

The system operates across multiple namespaces:

- `jozu-hub` - Main application components
- `jozu-workflows` - Argo Workflows and signing operations
- Custom namespaces for PostgreSQL and NATS when using external operators

### Release Process

The release orchestrator workflow automates on-premise releases by:

1. Cloning GitOps repository with current production state
2. Extracting SHA-based images from ECR using `scripts/extract-shas.sh`
3. Retagging and pushing to `images.jozu.com`
4. Updating `values.yaml` and `Chart.yaml`
5. Packaging and pushing Helm chart
6. Tagging microservice repositories with release version

This process ensures on-premise releases match the exact production state without manual intervention.

## Project Overview

KitOps is a CNCF open standards project for packaging, versioning, and securely sharing AI/ML projects. Built on the OCI standard, it serves as the reference implementation of the CNCF's ModelPack specification for vendor-neutral AI/ML interchange format.

## Development Commands

### Go CLI (Main Project)

```bash

# Build the CLI
go build -o kit .

# Run tests
go test ./...

# Run specific package tests
go test ./pkg/cmd/pack -v

# Install locally  
go install .
```

### Frontend Dev Mode UI

```bash
cd frontend/dev-mode
pnpm install
pnpm dev          # Development server
pnpm build        # Production build  
pnpm type-check   # TypeScript checking
pnpm lint         # ESLint with auto-fix
```

### Documentation Site

```bash
cd docs
pnpm install
pnpm docs:dev     # Development server
pnpm docs:build   # Production build
pnpm docs:preview # Preview built docs
```

## Architecture Overview

### Core CLI Structure

- **Entry point**: `main.go` → `cmd/root.go` using Cobra framework
- **Commands**: Individual CLI commands in `pkg/cmd/{command}/` (pack, unpack, push, pull, tag, list, inspect, etc.)
- **Libraries**: Core functionality in `pkg/lib/`:
  - `kitfile/`: Kitfile manifest parsing, validation, and generation
  - `repo/local/` & `repo/remote/`: Repository management for local storage and OCI registries
  - `harness/`: LLM integration with llamafile for local inference
  - `filesystem/`: File operations, caching, and ignore patterns
  - `network/`: Authentication and network operations

### Key Concepts

- **ModelKit**: Immutable OCI-compatible bundles containing code, models, datasets, and metadata
- **Kitfile**: YAML manifest format defining ModelKit contents (spec in `pkg/artifact/kitfile.md`)
- **Repository Management**: Uses OCI layout for local storage, compatible with container registries

### Frontend Components

- **Dev Mode UI**: Vue 3 + TypeScript SPA for local LLM interaction in `frontend/dev-mode/`
- **Documentation**: VitePress-based site in `docs/` with comprehensive guides and API reference

### CLI Command Pattern

All commands follow consistent structure:

- Command definition: `pkg/cmd/{command}/cmd.go`
- Main logic: `pkg/cmd/{command}/{command}.go`
- Options/configuration in separate files when complex

### Testing Structure

- Unit tests: `*_test.go` files alongside source
- Integration tests: `testing/` directory with comprehensive test data in `testing/testdata/`
- Test categories include manifest comparison, kitfile generation, pack/unpack workflows

### Output and Configuration

- All CLI output routed through `pkg/output/` for consistent formatting and logging
- Configuration handled via `pkg/lib/constants/` with environment variable support
- Caching managed in `pkg/lib/filesystem/cache/`
