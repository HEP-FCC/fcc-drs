# FCC Dataset Request System

A web-based dataset request tracking system for centrally produced datasets
used in analysis, detector design studies, and other physics related areas at
the [Future Circular Collider (FCC)](https://fcc.web.cern.ch/).

FCC-DRS keeps an eye on the dataset needs of the community around the FCC and makes
sure nothing falls through the cracks.

<p align="center"><img src="static/logo.png" alt="FCC Dataset Request System logo" width="160"/></p>

---

## Features

- **Submit requests** for FCC datasets with HEP-specific fields (dataset stage, use case, format, tags)
- **Track status** through the full pipeline: Draft → Pending Review → Approved → In Progress → Completed
- **Priority levels** — Low, Medium, High, Critical — with dashboard alerts for critical requests
- **Pipeline view** for coordinators: assignment, batch actions, inline status and priority overrides
- **Activity log** per request: comments, internal coordinator notes, system events with timestamps
- **Assignment** of requests to coordinators with dropdown selector
- **Batch actions** — approve, reject, complete, or move to in-progress across multiple requests at once
- **Filter & search** by status, priority, or free text
- **Bento-style dashboard** with live stats (total, pending, in-progress, completed)
- **Markdown + LaTeX math** rendering in titles and descriptions (KaTeX + marked.js)
- **Email notifications** on status changes (optional, via SMTP)
- **Dark / light / system** theme with persistent preference
- **Responsive design** — works on desktop and mobile

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

## Getting Started

### Prerequisites

- Go 1.22 or later

### Local development (dev mode)

CERN SSO is bypassed in dev mode. A simple form lets you pick any username and role — no credentials required.

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

No build step is required — Bulma is a pre-built CSS file.

The server starts on **http://localhost:5050**. The SQLite database is created automatically at `./data/requests.db` on first run (dev mode only).

### Email notifications (optional)

Set the following environment variables to enable email notifications on status changes:

```bash
export SMTP_HOST=smtp.cern.ch
export SMTP_PORT=587
export SMTP_USER=your-username
export SMTP_PASS=your-password
export SMTP_FROM=fcc-drs@cern.ch
```

If not set, email is silently disabled.

---

## Authentication & Roles

Authentication uses **CERN SSO** (Keycloak / OpenID Connect). Requester identity (name, username, email) is always taken from SSO and cannot be edited by users.

| Role          | Permissions |
|---------------|-------------|
| **Requester** | Submit requests, view all requests, edit own requests while draft or pending, add comments |
| **Coordinator** | Everything above + change status/priority on any request, assign requests, delete requests, batch actions, internal notes |
| **Admin**       | Everything above + manage user roles via the admin UI |

Role assignment:
- Roles are stored in the database and managed via the admin UI at `/admin/users`
- CERN usernames listed in `ADMIN_USERNAMES` receive the **admin** role on first login (bootstrap only)
- All other authenticated users receive the **requester** role by default

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
  base/              ← shared deployment, service, configmap
  overlays/
    staging/         ← staging namespace, hostname, image tag
    prod/            ← prod namespace, hostname, image tag, 2 replicas
```

### Prerequisites

- `oc login` access to both OpenShift projects (`fcc-drs-test`, `fcc-drs`)
- A PostgreSQL instance provisioned via CERN DBOD for each environment
- Two applications registered at the [CERN Application Portal](https://application-portal.web.cern.ch) (one per environment) to obtain OIDC client credentials

### 1. Build and push the container image

The production build uses the `prod` build tag, which enables PostgreSQL and disables SQLite.

```bash
docker build -t gitlab-registry.cern.ch/<your-group>/fcc-drs:<tag> .
docker push gitlab-registry.cern.ch/<your-group>/fcc-drs:<tag>
```

Update the `images.newName` and `images.newTag` fields in the relevant overlay's `kustomization.yaml` before deploying.

### 2. Fill in secrets

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

To bootstrap the first admin, edit `admin-usernames` in `openshift/base/configmap.yaml`:

```yaml
data:
  admin-usernames: "jsmith,adoe"
```

### 3. Deploy

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

### 4. Verify

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
| `ADMIN_USERNAMES`   | No  | Comma-separated CERN usernames granted admin role on first login |
| `PORT`              | No  | HTTP listen port (default `5050`) |
| `DEV_MODE`          | No  | Set to `TRUE` to bypass CERN SSO (local dev only) |
| `SQLITE_PATH`       | No  | Override SQLite file path (dev only, default `./data/requests.db`) |
| `SMTP_HOST`         | No  | SMTP server for email notifications |
| `SMTP_PORT`         | No  | SMTP port (default 587) |
| `SMTP_USER`         | No  | SMTP username |
| `SMTP_PASS`         | No  | SMTP password |
| `SMTP_FROM`         | No  | From address for notification emails |

---

## Dataset Request Fields

| Field                      | Required | Description |
|----------------------------|----------|-------------|
| Title                      | Yes | Short description; supports Markdown + LaTeX math |
| Description                | No  | Physics process, energy range, detector concept, selection criteria |
| Use Case                   | No  | Physics Analysis, Reconstruction Development, Detector Simulation, ML Training, ML Evaluation, Benchmarking, Calibration, Other |
| Dataset Stage              | No  | Generation, Simulation, Delphes, Reconstruction, Other |
| Working Group / Team       | No  | e.g. Tracker WG, Calorimetry WG |
| Format Needed              | Yes | EDM4hep, HepMC3, ROOT, … |
| Statistics / Estimated Size| Yes | Number of events or file size |
| Due Date                   | No  | When the data is needed |
| Priority                   | No  | Low / Medium / High / Critical |
| Tags                       | No  | e.g. `fcc-hh`, `fcc-ee`, `higgs`, `top`, `bsm`, `llp` |
| Notes                      | No  | Generator settings, beam conditions, special requirements |

Requester identity (name, username, email) is populated automatically from CERN SSO and is not editable.

---

## Contact & Support

- **General questions**: [FCC-PED-SoftwareAndComputing-MCProduction@cern.ch](mailto:FCC-PED-SoftwareAndComputing-MCProduction@cern.ch)
- **Platform issues & feature requests**: [github.com/HEP-FCC/fcc-drs/issues](https://github.com/HEP-FCC/fcc-drs/issues)

---

## Acknowledgements

Built with the assistance of [Claude](https://claude.ai) (Anthropic).
