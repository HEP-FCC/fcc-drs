# Coordinator guide

Coordinators review submitted requests and follow them through production.
Each coordinator belongs to one or more **coordinator groups**. A request is
assigned to a group based on the **Group / Team** chosen by the requester,
and the coordinators of that group are the ones notified about it.

## Reviewing a request

Every request needs two approvals:

- **Physics**: the request makes sense physics-wise.
- **Resources**: the computing resources can be provided.

Use the **Approve** and **Reject** buttons in the **Approvals** section of the request.

- When both are approved, the request moves to **Approved** automatically.
- When either is rejected, the request moves to **Rejected**.
- **Revert** clears a decision and moves the request back to **Under Review**.

If something is unclear, ask the requester in a comment before deciding.

## Managing a request

On the request page coordinators can also:

- change the **status** directly with the status selector,
- change the **priority** (Low, Medium, High, Critical),
- reassign the **coordinator group**,
- edit any section of the request, at any status,
- delete the request.

All changes are recorded in the **Activity** log.

## Internal notes

When commenting, tick **Internal (coordinators only)** to leave a note that the
requester cannot see. Internal notes are highlighted in the activity log.

## Campaigns

Approved requests are produced in **campaigns**. Campaigns are created and
closed by admins.

- Add an approved request to a campaign from its request page. The request then moves
  to **In Progress** automatically.
- The **Campaigns** page shows open and closed campaigns with their requests, and a
  **Needs Campaign Assignment** list of approved requests waiting to be scheduled.
- Requests in a **closed** campaign are locked: they cannot be removed or moved to another campaign.

When the production jobs are created, record their **Production IDs** (with an
optional label) on the request page. When the dataset is ready, set the status
to **Completed**, or to **Failed** if the production could not be finished.

## Batch actions

On the **Requests** page, select several requests with the checkboxes to approve,
reject, start or complete them at once.

## Email notifications

Coordinators of the assigned group (and admins) can receive emails about:

- new requests assigned to their group,
- status and approval changes,
- new comments, including internal notes.

Choose which of these you want on your **Profile** page.
