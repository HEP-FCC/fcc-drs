# Requester guide

## Creating a request

1. Click **New Request**.
2. Fill in the form. Fields marked with a red asterisk are required to submit:
   - **Title** and **Description**: the physics process, generator settings (see
     below), energy, detector concept, selection, etc.
   - **Group / Team**: the physics or software group the request belongs to. It decides who
     reviews your request.
   - **Use Case**, **Final Processing Stage** (Generation, Simulation, Delphes, Reconstruction,
     Other) and **Format**.
   - **Detector(s)**: required for every stage except Generation.
   - **Statistics** (number of events) and **Estimated Size**.
3. Optional fields: **Target Campaign**, **Key4hep Stack**, **Due Date**, **Priority**,
   **Additional Notes**, **Tags**, **Generator Cards** and **Related Requests**.
   Specify the **Key4hep Stack** only if you already know which one to target, e.g. when
   you use other datasets from a specific campaign.
4. Either save the request as a **Draft** or submit it straight away.

A submitted request gets the status **Under Review** and the MC Production contacts
of the selected group are notified.

## Who reviews your request

The MC Production contacts of the group review the request, may comment on it or
cancel it, and help with its technical preparation. A cancelled request is nothing
unusual. The physics approval is given by the MC Production contacts of the group,
the resources approval by the DRS admins.

## Drafts

Only the title is needed to save a draft. A draft can be edited freely and deleted.
When it is complete, open it and click **Submit Request**.

## Editing a request

You can edit your request while it is a **Draft** or **Under Review**. Click the
pencil icon next to a section to edit it in place. Once the request is approved,
only the MC Production contacts can change it. Ask them in a comment if something needs updating.

## Withdrawing or cancelling

While your request is **Under Review** you can:

- **Withdraw to Draft**: take it back for larger changes and submit it again later.
- **Cancel Request**: you no longer need the dataset.

## Cloning

**Clone** on any request opens a new form pre-filled with its content. The new
request is automatically linked to the original as a *variant*. This is the easiest
way to ask for a similar sample with different settings.

## Generator settings

Make the generator setup explicit in the description:

- matrix-element generator (e.g. Whizard, MadGraph5_aMC@NLO, Pythia 8),
- ISR treatment,
- parton shower,
- hadronization, including the tune (e.g. Pythia 8 `Tune:ee`).

Missing settings are clarified with the MC Production contacts before the
production starts.

## Generator cards

Every request should ideally come with its generator cards attached (e.g. `.py`,
`.cmnd`, `.dat`), with a note in the description on where you got them from.

The recommended source is the main branch of
[FCC-config](https://github.com/HEP-FCC/FCC-config), which keeps the reference
cards. Start from one of them and list any modifications in the description.

Only plain-text files are accepted, up to 1 MB each. They can be viewed or downloaded
from the request page.

## Related requests

Link your request to other requests as *Extends*, *Depends on*, *Variant of* or
*Related to*. Writing a request key, e.g. `DRS-42`, in the description or notes
creates a *mention* link automatically.

## Following your request

- The **Dashboard** shows overall statistics and the most recent requests.
- **Requests** lists all requests, with filters by status and priority and a free-text search.
- The **Activity** section of a request shows comments and every status change.
  Use it to answer questions from the MC Production contacts or to ask about the progress.
- **Approvals** show the state of the physics and resources review.
- Once the request is in a campaign, the **Transformation IDs** of the production are listed.

## Finding the produced datasets

Produced datasets are listed in [DaMI](https://fcc-dami.web.cern.ch/). For now they
are stored on EOS under `/eos/experiment/fcc/prod`. The location will be fully
specified once the datasets are distributed across all storage elements of the FCC VO.

## Email notifications

By default you receive emails about:

- status and approval changes on your requests,
- new comments on your requests.

You can turn them off on your **Profile** page. You can also set a preferred display
name and upload an avatar there.
