---
title: Web architecture
type: reference
description: Current structure of the web frontend — Feature-Sliced Design layers.
resource: https://github.com/gavelcode/gavel/tree/main/apps/web
tags: [architecture, web, frontend]
---

# Web Architecture

> This documents the current web app structure. It is a working implementation,
> not necessarily the final design.

Scope: `apps/web/` — browser dashboard consuming the server HTTP API.

## Done

Feature-Sliced Design layers (app → pages → features → entities → shared).
Each entity is self-contained (model + api + mappers + ui). Pages compose
from entities and shared. MSW for API mocking in tests. No DDD ceremony.

---

## Role

The web app (`apps/web/`) is a browser dashboard for projects, case files,
findings, quality gates, and administration. It consumes the server's HTTP API.

## Stack

React + TypeScript + Vite. TailwindCSS for styling. TanStack Query for
server state. Vitest + MSW for testing.

## Structure (Feature-Sliced Design)

```
apps/web/src/
├── app/                       # App shell
│   ├── App.tsx               # Root component
│   ├── router.tsx            # Route definitions
│   ├── providers.tsx         # Global providers (query client, auth, theme)
│   ├── error-boundary.tsx    # Top-level error boundary
│   └── layout/              # Sidebar, protected route wrapper
├── pages/                     # Route-level page components
│   ├── dashboard.tsx
│   ├── projects.tsx
│   ├── project-detail.tsx
│   ├── casefiles.tsx
│   ├── casefile-detail.tsx
│   ├── issues.tsx
│   ├── login.tsx
│   ├── admin-users.tsx
│   ├── tokens.tsx
│   └── ...
├── features/                  # Feature-scoped components
│   └── search/               # Command palette
├── entities/                  # Shared entity types, API calls, UI components
│   ├── casefile/             # model, api, mappers, ui/
│   ├── finding/
│   ├── project/
│   ├── gavelspace/
│   ├── pull-request/
│   ├── token/
│   ├── user/                 # auth-provider, use-auth
│   └── search/
├── shared/                    # Cross-cutting utilities
│   ├── api/                  # HTTP client, shared types
│   ├── auth/                 # Protected route component
│   ├── lib/                  # Formatting, theme, utils
│   └── ui/                   # Design primitives (badge, button, card, etc.)
└── test/                      # MSW handlers, test setup, render helpers
```

## Key decisions

- **Feature-Sliced Design, not DDD.** The web app uses FSD layers
  (`app → pages → features → entities → shared`) instead of DDD layers.
  DDD ceremony does not add value in a frontend that is essentially a
  read-heavy API consumer.
- **Entity = model + api + mappers + ui.** Each entity folder is self-contained:
  TypeScript types, API fetch functions, server-to-UI mappers, and reusable
  UI components for that entity.
- **Pages own layout and composition.** A page imports from entities and shared,
  composes them, and handles route-specific logic.
- **MSW for API mocking in tests.** No direct fetch mocking — MSW intercepts
  at the network level for realistic tests.
