# FCC Dataset Request System

FCC-DRS collects and tracks requests for centrally produced datasets used in
physics analyses, detector design studies, and software development at the FCC.
Every request goes through a review and is followed until the dataset is produced.

Log in with your CERN SSO account. New users start as **requesters**.

## Roles

| Role | What you can do |
|------|-----------------|
| **Requester** | Submit dataset requests, edit them while they are under review, follow their progress and discuss them with coordinators. |
| **Coordinator** | Review requests, give physics and resources approvals, assign groups, priorities and campaigns, and track production. |
| **Admin** | Everything a coordinator can do, plus managing users, coordinator groups and campaigns. |

If you need coordinator access or have questions about FCC-DRS, contact
[FCC-PED-SoftwareAndComputing-MCProduction@cern.ch](mailto:FCC-PED-SoftwareAndComputing-MCProduction@cern.ch).

## Request lifecycle

```
Draft → Under Review → Approved → In Progress → Completed
```

| Status | Meaning |
|--------|---------|
| **Draft** | Saved but not yet submitted for review; you can still edit or delete it. |
| **Under Review** | Submitted and waiting for the physics and resources approvals. |
| **Approved** | Both approvals granted; waiting to be scheduled in a production campaign. |
| **In Progress** | Assigned to a production campaign and being produced. |
| **Completed** | The dataset has been produced. |
| **Rejected** | One of the approvals was denied. |
| **Failed** | The production could not be completed. |
| **Cancelled** | Withdrawn by the requester or closed by a coordinator. |

## Where to go next

- **Requester** tab: how to submit and follow a request.
- **Coordinator** tab: how requests are reviewed and scheduled.

## Formatting

Descriptions, notes and comments support **Markdown** and LaTeX math
(`$...$` inline, `$$...$$` for display equations). Writing `#` followed by a
request number links to that request.
