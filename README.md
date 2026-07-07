# FCC Dataset Request System

A web-based dataset request tracking system for centrally produced datasets
used in analysis, detector design studies, and other physics related areas at
the [Future Circular Collider (FCC)](https://fcc.web.cern.ch/).

FCC-DRS keeps an eye on the dataset needs of the community around the FCC and makes
sure nothing falls through the cracks.

**Production**: [fcc-drs.web.cern.ch](https://fcc-drs.web.cern.ch) | **Staging**: [fcc-drs-test.web.cern.ch](https://fcc-drs-test.web.cern.ch)

<p align="center"><img src="static/logo.png" alt="FCC Dataset Request System logo" width="160"/></p>

---

## Features

- **Submit and track requests** through the full pipeline: Draft → Pending Review → Approved → In Progress → Completed
- **Role-based access** — requesters, coordinators, and admins with appropriate permissions at each stage
- **Activity log** per request: comments, internal notes, and system events with timestamps
- **Filter & search** by status, priority, or free text; Markdown + LaTeX math in titles and descriptions
- **Email notifications** on status changes (optional, via SMTP)

---

## Tech Stack

| Layer     | Technology                                                              |
|-----------|-------------------------------------------------------------------------|
| Backend   | Go 1.22+ (`net/http` standard library)                                  |
| Frontend  | HTMX 2 + Bulma 1.0 CSS (self-hosted, pre-built)                         |
| Database  | SQLite (local dev, no CGO) / PostgreSQL via CERN DBOD (production)      |
| Auth      | CERN SSO via OpenID Connect (Keycloak)                                  |
| Math/MD   | KaTeX + marked.js (self-hosted)                                         |
| Fonts     | Inter (self-hosted woff2 subsets)                                       |
| Email     | Go stdlib `net/smtp`                                                    |

Single binary, no CGO required. No external CDN dependencies at runtime.

---

## Local Development

Go 1.22 or later required. CERN SSO is bypassed in dev mode — a simple form lets you pick any username and role.

```bash
git clone https://github.com/HEP-FCC/fcc-drs
cd fcc-drs
DEV_MODE=TRUE go run ./cmd/fcc-drs
```

Open **http://localhost:5050**, enter a username, choose a role (requester, coordinator, or admin), and log in.

### Build

```bash
make build      # production build (PostgreSQL, version from git tag)
make build-dev  # development build (SQLite, version = "dev")
make reseed     # drop and recreate the dev DB, apply scripts/seed.sql, then exit
make run        # start in dev mode (DEV_MODE=TRUE, SQLite)
```

The production build injects the current git tag as the version string shown in the footer via `-ldflags`. Omitting it (or building with `go build` directly) defaults to `dev`.

### Front-end assets

All JS/CSS dependencies (HTMX, Bulma CSS, KaTeX, marked, Inter font) are self-hosted under `static/vendor/`. Run once after cloning:

```bash
make assets
```


---

## Authentication & Roles

Authentication uses **CERN SSO** (Keycloak / OpenID Connect). Requester identity (name, username, email) is always taken from SSO and cannot be edited by users.

| Role          | Permissions |
|---------------|-------------|
| **Requester** | Submit requests, view all requests, edit own requests while draft or pending, add comments |
| **Coordinator** | Everything above + change status/priority on any request, assign requests, delete requests, batch actions, internal notes |
| **Admin**       | Everything above + manage user roles via the admin UI |

Role assignment:
- All new users receive the **requester** role on first login
- Roles are managed via the admin UI at `/admin/users`
- **Dev mode**: select any role from the login form — no bootstrap needed
- **Staging/Production**: to bootstrap the first admin, update the database directly after first login:
  ```sql
  UPDATE users SET role = 'admin' WHERE username = '<cern-username>';
  ```

---

## Deployment (CERN PaaS / OpenShift)

FCC-DRS runs on [CERN PaaS](https://paas.cern.ch) (OpenShift) with a PostgreSQL database provided by the [CERN DBOD](https://dbod.web.cern.ch) service.

There are two deployed environments:

| Environment | URL | Namespace |
|-------------|-----|-----------|
| Staging | `fcc-drs-test.web.cern.ch` | `fcc-drs-test` |
| Production | `fcc-drs.web.cern.ch` | `fcc-drs` |

Manifests are managed with [Kustomize](https://kustomize.io) (built into `oc`/`kubectl`):

```
openshift/
  base/              ← shared deployment, service
  overlays/
    staging/         ← staging namespace, hostname, image tag
    prod/            ← prod namespace, hostname, image tag, 2 replicas
```

### Prerequisites

- `oc login` access to both OpenShift projects (`fcc-drs-test`, `fcc-drs`)
- A PostgreSQL instance provisioned via CERN DBOD for each environment
- Two applications registered at the [CERN Application Portal](https://application-portal.web.cern.ch) (one per environment) to obtain OIDC client credentials

### 1. Fill in secrets

Copy the example secret for each environment, fill in real credentials, and apply it. The `secret.yaml` files are gitignored and must never be committed with actual values.

```bash
cp openshift/overlays/staging/secret.example.yaml openshift/overlays/staging/secret.yaml
cp openshift/overlays/prod/secret.example.yaml    openshift/overlays/prod/secret.yaml
```

```yaml
stringData:
  database-url: "postgresql://user:password@dbod-host.cern.ch:5432/database?sslmode=require"
  oidc-client-id: "your-client-id"
  oidc-client-secret: "your-client-secret"
  oidc-redirect-url: "https://<hostname>/auth/callback"
```

### 2. Deploy

```bash
# Staging
make deploy-staging

# Production
make deploy-prod
```

These apply the secret first, then the full Kustomize overlay. Equivalent to:

```bash
oc apply -f openshift/overlays/<env>/secret.yaml
oc apply -k openshift/overlays/<env>
```

### 3. Verify

```bash
oc get pods        # pod should reach Running state
oc get route fcc-drs   # shows the public URL
oc logs -f deployment/fcc-drs  # tail logs
```

The database schema is created automatically on first startup. No manual migration step is required.

### Environment variables reference

| Variable            | Required | Description |
|---------------------|----------|-------------|
| `DATABASE_URL`      | Yes (prod) | PostgreSQL connection string from CERN DBOD |
| `OIDC_CLIENT_ID`    | Yes (prod) | CERN Application Portal client ID |
| `OIDC_CLIENT_SECRET`| Yes (prod) | CERN Application Portal client secret |
| `OIDC_REDIRECT_URL` | Yes (prod) | Must be `https://<hostname>/auth/callback` |
| `PORT`              | No  | HTTP listen port (default `5050`) |
| `DEV_MODE`          | No  | Set to `TRUE` to bypass CERN SSO (local dev only) |
| `SQLITE_PATH`       | No  | Override SQLite file path (dev only, default `./data/requests.db`) |
| `SMTP_HOST`         | No  | SMTP server for email notifications |
| `SMTP_PORT`         | No  | SMTP port (default 587) |
| `SMTP_USER`         | No  | SMTP username |
| `SMTP_PASS`         | No  | SMTP password |
| `SMTP_FROM`         | No  | From address for notification emails |

---

## Contact & Support

- **General questions**: [FCC-PED-SoftwareAndComputing-MCProduction@cern.ch](mailto:FCC-PED-SoftwareAndComputing-MCProduction@cern.ch)
- **Platform issues & feature requests**: [github.com/HEP-FCC/fcc-drs/issues](https://github.com/HEP-FCC/fcc-drs/issues)

---

## Acknowledgements

Built with the assistance of [Claude](https://claude.ai) (Anthropic).
