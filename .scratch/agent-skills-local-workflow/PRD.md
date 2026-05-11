# PRD: Agent Skills Local Markdown Workflow

Status: needs-triage
Created: 2026-05-07

## Problem Statement

The MockServer repo is a local-first project with no configured remote issue tracker. The maintainer wants to use engineering skills such as `to-prd`, `to-issues`, `triage`, `diagnose`, `tdd`, and `improve-codebase-architecture`, but those skills need a consistent repository-level contract for where PRDs and issues are published, how triage states are represented, and where domain documentation or architectural decisions should be read from.

Without that contract, each skill would need to rediscover the workflow on every run, risk inventing different file layouts, and risk mixing task planning artifacts into unrelated source changes. The maintainer also needs the setup to fit the current greenfield posture of the project: prefer a clean, direct workflow over compatibility layers or remote tracker assumptions that do not exist yet.

## Solution

Provide a repository-local agent workflow that uses markdown files as the issue tracker, applies the default five-role triage vocabulary, and treats the repo as a single-context project for domain documentation. The workflow should let PRDs and future implementation issues be published under a local scratch workspace while keeping the durable instructions discoverable through the repo's agent instruction file and agent-specific documentation.

The result should make the following behavior predictable:

- PRDs are published as local markdown documents.
- Implementation issues can later be generated under the same feature workspace.
- Triage state is stored directly in markdown using the default `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, and `wontfix` states.
- Domain-aware skills read the repo-level domain glossary and architectural decision records when they exist, but do not require them to be created upfront.
- The workflow remains lightweight enough for a solo/local project and can later be replaced by GitHub, GitLab, or another issue tracker if the repo starts using one.

## User Stories

1. As the repo maintainer, I want PRDs to be created inside the repo, so that planning artifacts stay next to the code they describe.
2. As the repo maintainer, I want local markdown to be the issue tracker, so that I can use planning skills before setting up GitHub or GitLab Issues.
3. As the repo maintainer, I want each feature to have its own local workspace, so that PRDs, implementation issues, and comments for the same feature are grouped together.
4. As the repo maintainer, I want generated PRDs to start in `needs-triage`, so that every new planning artifact enters a clear review state.
5. As the repo maintainer, I want the default triage labels to be used unchanged, so that skills can interoperate without translating custom status names.
6. As the repo maintainer, I want the local issue convention documented, so that future agents do not need to ask where to publish issues.
7. As the repo maintainer, I want the triage label mapping documented, so that future agents do not create duplicate or inconsistent labels.
8. As the repo maintainer, I want the domain documentation layout documented, so that future agents know where to look for project language and decisions.
9. As the repo maintainer, I want missing domain docs to be treated as normal, so that the project does not need placeholder `CONTEXT` or ADR files before they are useful.
10. As the repo maintainer, I want architectural decisions to be captured as ADRs later, so that important design choices can be reused by future diagnosis and refactoring work.
11. As the repo maintainer, I want the setup to avoid compatibility work, so that this greenfield project can keep the simplest correct workflow.
12. As the repo maintainer, I want the generated workflow files to be small and readable, so that I can adjust them manually when the project process changes.
13. As the repo maintainer, I want future `to-issues` runs to know where implementation issues belong, so that PRD-to-ticket conversion is deterministic.
14. As the repo maintainer, I want future `triage` runs to know which statuses are valid, so that issue state transitions are consistent.
15. As the repo maintainer, I want future `diagnose` runs to respect domain docs and ADRs, so that bug investigation uses project language and avoids reopening settled decisions.
16. As the repo maintainer, I want future `tdd` runs to respect domain docs and ADRs, so that tests describe behavior using the repo's vocabulary.
17. As the repo maintainer, I want future architecture-improvement runs to read ADRs first, so that refactor proposals account for existing decisions.
18. As a coding agent, I want a single instruction surface that points me to the issue tracker, label vocabulary, and domain docs, so that I can start work with less repo-specific guesswork.
19. As a coding agent, I want local markdown publishing rules, so that I can create PRDs and issues without network access or external CLI authentication.
20. As a coding agent, I want comments and discussion history to append to the local issue document, so that decisions are preserved with the task.
21. As a coding agent, I want the repo to define whether it is single-context or multi-context, so that I read the correct domain documentation before modifying code.
22. As a future contributor, I want the workflow to be visible in repository instructions, so that I can understand how agent-driven tasks are organized.
23. As a future contributor, I want PRDs and issues to be separated from source code modules, so that planning artifacts do not obscure runtime, frontend, or database implementation files.
24. As a future contributor, I want the workflow to be easy to migrate later, so that moving from local markdown to GitHub or another tracker does not require rewriting feature intent.
25. As a future reviewer, I want PRDs to carry their triage status near the top, so that I can quickly see whether a proposal has been evaluated.
26. As a future reviewer, I want implementation details to avoid stale file-level commitments where possible, so that PRDs remain valid after normal refactors.
27. As a future reviewer, I want testing decisions to focus on external behavior, so that generated implementation work does not overfit to current internals.
28. As the repo maintainer, I want this workflow to coexist with existing uncommitted code changes, so that setup can be committed independently from unrelated feature work.

## Implementation Decisions

- The issue tracker is local markdown rather than GitHub, GitLab, Jira, Linear, or another external system.
- A feature-oriented local workspace is the unit of organization for PRDs and implementation issues.
- PRDs enter the workflow with the `needs-triage` status.
- The canonical triage roles use the default names unchanged: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, and `wontfix`.
- The repo is treated as a single-context project for domain documentation.
- Domain-aware skills should read the repo-level glossary and architectural decision records when they exist.
- Missing glossary or ADR files are not setup failures. Skills should continue silently until a producer workflow creates those docs from real decisions.
- The agent instruction surface should summarize the issue tracker, triage labels, and domain docs, then point to dedicated documentation for details.
- The dedicated agent documentation should be the source of truth for future skills, so changes to workflow conventions can be made without rewriting every skill prompt.
- The setup should not create placeholder domain docs or ADRs before there is real domain language or a real architectural decision to record.
- The setup should not modify MockServer runtime behavior, frontend behavior, database schema, API contracts, rule matching, publishing, rollback, metrics, or admin authentication.
- Future migration to an external issue tracker should be handled by changing the tracker documentation and instruction summary, not by preserving a compatibility adapter.

## Testing Decisions

- Good tests for this workflow verify externally visible behavior: a PRD can be published to the configured local tracker, a triage status is present near the top, and future skills can discover the tracker and label conventions from documented instructions.
- Tests should not assert implementation details of individual skills, prompt wording, or internal command sequences.
- The agent instruction surface should be inspected to confirm it contains issue tracker, triage label, and domain docs summaries.
- The local issue tracker documentation should be inspected to confirm it defines feature workspaces, PRD location, issue numbering, triage status placement, and comment placement.
- The triage label documentation should be inspected to confirm all five canonical roles map to the default labels.
- The domain documentation should be inspected to confirm the repo is single-context and that missing glossary or ADR files are acceptable.
- A whitespace check should be run after editing markdown files to catch trailing whitespace or patch hygiene problems.
- Future higher-level tests can be added around generated PRDs and issues by invoking the publishing skills and asserting the expected markdown files and status lines exist.

## Out of Scope

- Creating GitHub, GitLab, Jira, Linear, or other external issue tracker integration.
- Creating or migrating remote repository configuration.
- Implementing new behavior in MockServer runtime, rule matching, admin APIs, frontend pages, persistence, metrics, or authentication.
- Creating placeholder `CONTEXT` documents without real glossary content.
- Creating placeholder ADRs without a concrete architectural decision.
- Converting existing uncommitted source changes into issues.
- Committing the setup or any generated PRD to Git.
- Defining a complete project management process beyond the local PRD, issue, triage, and domain-doc conventions needed by the engineering skills.

## Further Notes

This PRD was generated from the current setup conversation. The repo currently uses a local-first workflow and has no configured Git remote. The setup is intentionally lightweight: the local markdown tracker is sufficient for immediate PRD and issue generation, while the documented conventions leave room to switch to a remote tracker later if the project starts using one.
