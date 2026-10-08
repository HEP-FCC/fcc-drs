# Requester guide

## Creating a request

1. Click **New Request**.
2. Fill in the form. Fields marked with a red asterisk are required to submit:
   - **Title** and **Description**: the physics process, energy, detector concept, selection, etc.
   - **Group / Team**: the physics or software group the request belongs to. It decides which
     coordinator group reviews your request.
   - **Use Case**, **Final Processing Stage** (Generation, Simulation, Delphes, Reconstruction)
     and **Format**.
   - **Detector(s)**: required for every stage except Generation.
   - **Statistics** (number of events) and **Estimated Size**.
3. Optional fields: **Target Campaign**, **Key4hep Stack**, **Due Date**, **Priority**,
   **Additional Notes**, **Tags**, **Generator Cards** and **Related Requests**.
4. Either save the request as a **Draft** or submit it straight away.

A submitted request gets the status **Under Review** and the coordinators of the
selected group are notified.

## Drafts

Only the title is needed to save a draft. A draft can be edited freely and deleted.
When it is complete, open it and click **Submit Request**.

## Editing a request

You can edit your request while it is a **Draft** or **Under Review**. Click the
pencil icon next to a section to edit it in place. Once the request is approved,
only coordinators can change it. Ask them in a comment if something needs updating.

## Withdrawing or cancelling

While your request is **Under Review** you can:

- **Withdraw to Draft**: take it back for larger changes and submit it again later.
- **Cancel Request**: you no longer need the dataset.

## Cloning

**Clone** on any request opens a new form pre-filled with its content. The new
request is automatically linked to the original as a *variant*. This is the easiest
way to ask for a similar sample with different settings.

## Generator cards

Attach generator configuration files (e.g. `.py`, `.cmnd`, `.dat`) to the request.
Only plain-text files are accepted, up to 1 MB each. They can be viewed or downloaded
from the request page.

## Related requests

Link your request to other requests as *Extends*, *Depends on*, *Variant of* or
*Related to*. Writing `#` followed by a request number in the description or notes
creates a *mention* link automatically.

## Following your request

- The **Dashboard** shows overall statistics and the most recent requests.
- **Requests** lists all requests, with filters by status and priority and a free-text search.
- The **Activity** section of a request shows comments and every status change.
  Use it to answer questions from coordinators or to ask about the progress.
- **Approvals** show the state of the physics and resources review.
- Once the request is in a campaign, the **Production IDs** of the production jobs are listed.

## Email notifications

On your **Profile** page you can choose to receive emails about:

- status and approval changes on your requests,
- new comments on your requests.

You can also set a preferred display name and upload an avatar there.
