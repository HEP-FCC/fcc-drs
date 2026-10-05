-- Seed data: users, campaigns, dataset requests, and examples of every
-- feature that attaches to a request (production IDs, activity/comments,
-- relations, a generator card).
-- Run with: sqlite3 data/requests.db < scripts/seed.sql

INSERT OR IGNORE INTO users (username, display_name, email, role) VALUES
  ('akowalski',  'Anna Kowalski',    'akowalski@cern.ch',  'requester'),
  ('mrossi',     'Marco Rossi',      'mrossi@cern.ch',     'requester'),
  ('lchen',      'Li Chen',          'lchen@cern.ch',      'requester'),
  ('hmueller',   'Hans Mueller',     'hmueller@cern.ch',   'requester'),
  ('sfernandez', 'Sofia Fernandez',  'sfernandez@cern.ch', 'requester'),
  ('tdupont',    'Thomas Dupont',    'tdupont@cern.ch',    'requester'),
  ('epatel',     'Elena Patel',      'epatel@cern.ch',     'coordinator'),
  ('rkim',       'Robert Kim',       'rkim@cern.ch',       'coordinator');

INSERT OR IGNORE INTO campaigns (name, status, tag) VALUES
  ('Spring2026', 'open',   'spring2026'),
  ('Summer2026', 'open',   'summer2026'),
  ('Winter2025', 'closed', 'winter2025');

INSERT INTO dataset_requests
  (title, description, requester_name, requester_username, requester_email,
   assigned_group_id, campaign_id, dataset_type, use_case, status, priority,
   estimated_size, statistics, target_campaign, key4hep_stack, detector,
   format, due_date, notes, tags,
   physics_approval, resources_approval,
   created_at, updated_at)
VALUES

-- 1: in_progress, campaign-assigned, single detector, production IDs, rich comment thread
('Higgs → bb̄ signal at FCC-ee √s=240 GeV',
 'Full simulation of $e^+e^- \to ZH, H \to b\bar{b}$ at 240 GeV centre-of-mass energy. Need at least 10M events for signal efficiency studies.

See also #13 for the complementary single-top sample used as a cross-check background.',
 'Anna Kowalski', 'akowalski', 'akowalski@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Higgs physics'),
 (SELECT id FROM campaigns WHERE name = 'Spring2026'),
 'simulation', 'physics_analysis', 'in_progress', 'critical',
 '~2 TB', '10M events', 'Spring2026', 'key4hep-2025-11-28', 'IDEA',
 'EDM4hep', '2026-06-30',
 'Pythia8 + full Geant4 simulation. Require full track + cluster truth links.',
 'fcc-ee,higgs,bbar',
 'approved', 'approved',
 '2026-01-15 10:00:00', '2026-03-10 14:22:00'),

-- 2: completed, generation stage — no detector required
('$t\bar{t}$ production at FCC-hh 100 TeV',
 'Inclusive top-quark pair production for detector acceptance studies. Needed as a SM benchmark before BSM search campaigns begin.',
 'Marco Rossi', 'mrossi', 'mrossi@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'FCC-hh physics'), NULL,
 'generation', 'physics_analysis', 'completed', 'high',
 '~5 TB', '50M events', '', '', '',
 'HepMC3', '2025-12-01',
 'MadGraph5 + Pythia8 shower. No detector sim needed at this stage.',
 'fcc-hh,top',
 'approved', 'approved',
 '2025-09-01 08:30:00', '2026-01-20 09:00:00'),

-- 3: in_progress, with an internal coordinator note
('Drell-Yan background for LLP search',
 'Large statistics DY sample for background estimation in the long-lived particle displaced vertex analysis. Need accurate modelling of low-pT tracks.',
 'Li Chen', 'lchen', 'lchen@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'BSM physics'), NULL,
 'simulation', 'physics_analysis', 'in_progress', 'high',
 '~8 TB', '100M events', 'Summer2026', 'key4hep-2025-11-28', 'IDEA',
 'EDM4hep', '2026-07-15',
 'Sherpa 2.2.15. Require QCD multi-jet merging up to 4 jets.',
 'fcc-hh,llp,background',
 'approved', '',
 '2026-02-01 11:00:00', '2026-04-05 16:30:00'),

-- 4: pending, not yet approved — no detector shown since stage pending approval doesn't block drafting
('Z-pole precision: $Z \to \mu\mu$ high-statistics sample',
 'Dedicated $Z \to \mu^+\mu^-$ sample at $\sqrt{s}=91.2$ GeV for electroweak precision observables. Requires radiative corrections at NNLO.',
 'Sofia Fernandez', 'sfernandez', 'sfernandez@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Electroweak physics'), NULL,
 'simulation', 'physics_analysis', 'pending', 'critical',
 '~50 TB', '1B events', '', '', 'IDEA',
 'EDM4hep', '2026-09-01',
 'KKMC generator mandatory. Full detector simulation.',
 'fcc-ee,zpole,ewk,precision',
 '', '',
 '2026-03-20 09:15:00', '2026-03-20 09:15:00'),

-- 5: approved, multi-detector example
('Tracker alignment sample: $e^+e^- \to \mu\mu$',
 'Clean dimuon events for inner tracker alignment studies. Flat pT spectrum from 10 to 100 GeV needed. Run through both detector concepts so the alignment procedure can be cross-validated.

See #13 for the fast-sim comparison that reuses this generator-level sample.',
 'Hans Mueller', 'hmueller', 'hmueller@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Geometry, Simulation'), NULL,
 'simulation', 'calibration', 'approved', 'medium',
 '~200 GB', '5M events', 'Spring2026', 'key4hep-2025-11-28', 'IDEA, CLD',
 'EDM4hep', '2026-05-31',
 'Uniform angular distribution.',
 'fcc-ee,tracker,alignment,calibration',
 'approved', 'approved',
 '2026-02-10 13:00:00', '2026-04-01 10:00:00'),

-- 6: pending
('WW threshold scan: $e^+e^- \to W^+W^-$',
 'Events at 6 energy points around $\sqrt{s}=161$ GeV for W-mass measurement. Each energy point needs equal statistics.',
 'Thomas Dupont', 'tdupont', 'tdupont@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Electroweak physics'), NULL,
 'simulation', 'physics_analysis', 'pending', 'high',
 '~3 TB total', '30M events per energy point', '', '', 'IDEA',
 'EDM4hep', '2026-08-01',
 'Whizard generator. Include ISR and beam energy spread.',
 'fcc-ee,wmass,ww,threshold',
 '', '',
 '2026-03-05 10:00:00', '2026-03-05 10:00:00'),

-- 7: completed, generator card attached
('ECAL single-particle response: electrons 1–100 GeV',
 'Geant4 single-electron gun for calorimeter response calibration. Log-spaced energy points needed.',
 'Anna Kowalski', 'akowalski', 'akowalski@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Geometry, Simulation'), NULL,
 'simulation', 'calibration', 'completed', 'medium',
 '~500 GB', '500k events per energy point', '', 'key4hep-2025-03-14', 'IDEA',
 'EDM4hep', '2025-11-30',
 'Full Geant4 — no fast sim. Noble-liquid ECAL geometry. Particle gun config attached below.',
 'fcc-ee,ecal,calibration,single-particle',
 'approved', 'approved',
 '2025-08-15 09:00:00', '2025-12-01 11:00:00'),

-- 8: pending, generation stage, related to #21 below
('Gluino pair production: $\tilde{g}\tilde{g} \to t\bar{t}\tilde{\chi}^0_1\tilde{\chi}^0_1$',
 'SUSY gluino pair production grid scan over $(m_{\tilde{g}}, m_{\tilde{\chi}})$ parameter space for FCC-hh prospect paper.',
 'Li Chen', 'lchen', 'lchen@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'BSM physics'), NULL,
 'generation', 'physics_analysis', 'pending', 'medium',
 '~1 TB total', '50k events per grid point, ~200 points', '', '', '',
 'HepMC3', '2026-10-01',
 'MadGraph5_aMC@NLO + Pythia8 + fastjet. Provide cross-section table.',
 'fcc-hh,susy,bsm,gluino',
 '', '',
 '2026-04-01 14:00:00', '2026-04-01 14:00:00'),

-- 9: in_progress, campaign-assigned, production ID, markdown-rich comment (table)
('ML training sample: jet flavour tagging',
 'Large mixed-flavour jet sample (b, c, light, gluon) for training a graph-neural-network flavour tagger.',
 'Marco Rossi', 'mrossi', 'mrossi@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'High-level reconstruction'),
 (SELECT id FROM campaigns WHERE name = 'Summer2026'),
 'reconstruction', 'ml_training', 'in_progress', 'high',
 '~10 TB', '200M jets', 'Summer2026', 'key4hep-2025-11-28', 'IDEA',
 'EDM4hep', '2026-06-15',
 'Balanced class weights required. Include pile-up at $\mu=200$.',
 'fcc-hh,ml,jets,flavour-tagging',
 'approved', 'approved',
 '2026-01-20 10:00:00', '2026-03-25 17:00:00'),

-- 10: draft — detector not required yet since stage not yet submitted
('Lepton flavour violation: $H \to e\mu$',
 'LFV Higgs decay search sample. Signal process $e^+e^- \to ZH, H \to e\mu$ plus irreducible backgrounds.',
 'Sofia Fernandez', 'sfernandez', 'sfernandez@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Higgs physics'), NULL,
 'simulation', 'physics_analysis', 'draft', 'medium',
 '', '2M signal + 20M background events', '', '', '',
 'EDM4hep', '2026-11-01',
 'Need to agree on background composition with Higgs WG convener before production.',
 'fcc-ee,higgs,lfv,bsm',
 '', '',
 '2026-04-10 15:00:00', '2026-04-10 15:00:00'),

-- 11: failed — showcases the failed status
('PID benchmark: pions and kaons in RICH',
 'Single-particle gun (π±, K±) across full RICH acceptance for PID efficiency vs. momentum parametrisation.',
 'Hans Mueller', 'hmueller', 'hmueller@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Geometry, Simulation'),
 (SELECT id FROM campaigns WHERE name = 'Winter2025'),
 'simulation', 'benchmarking', 'failed', 'low',
 '', '1M events per species per momentum point', '', 'key4hep-2025-03-14', 'IDEA',
 'ROOT', '2025-10-15',
 'Dual-RICH geometry. Production failed partway through — see activity log for the DIRAC job failure note.',
 'fcc-ee,pid,rich,benchmark',
 'approved', 'approved',
 '2025-06-01 09:00:00', '2025-10-20 09:00:00'),

-- 12: rejected
('Inclusive $b\bar{b}$ at FCC-ee for B-physics',
 'Large $e^+e^- \to b\bar{b}$ sample at $\sqrt{s}=10.58$ GeV (Υ(4S) region) for B-meson oscillation studies.',
 'Thomas Dupont', 'tdupont', 'tdupont@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Flavour physics'), NULL,
 'simulation', 'physics_analysis', 'rejected', 'medium',
 '~20 TB', '500M events', '', '', 'IDEA',
 'EDM4hep', '2026-07-01',
 'Rejected: FCC-ee Υ(4S) running not in the current baseline. Re-submit for Z-pole or 240 GeV.',
 'fcc-ee,bphysics,bbar',
 'rejected', '',
 '2026-01-05 10:00:00', '2026-02-14 11:00:00'),

-- 13: draft, relation back to #1 and #5
('Single-top associated production at FCC-ee 365 GeV',
 'Simulation of $e^+e^- \to tW^-\bar{b} + \text{c.c.}$ near the top threshold for anomalous $Wtb$ coupling studies. Complementary to #1 and reuses the generator setup from #5.',
 'Anna Kowalski', 'akowalski', 'akowalski@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Top-quark physics'), NULL,
 'simulation', 'physics_analysis', 'draft', 'low',
 '~300 GB', '2M events', '', '', '',
 'EDM4hep', '2026-12-01',
 'Whizard with ISR. Both semileptonic and hadronic W decay modes needed.',
 'fcc-ee,top,single-top,wtb',
 '', '',
 '2026-04-02 09:00:00', '2026-04-02 09:00:00'),

-- 14: in_progress, multi-detector, campaign + production ID
('Track reconstruction benchmarking: $t\bar{t}$ at 365 GeV',
 'Input sample for tracking algorithm benchmarking (Acts). Dense environment with high track multiplicity. Run through both detector concepts under consideration.',
 'Marco Rossi', 'mrossi', 'mrossi@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Digitization and Reconstruction Software'),
 (SELECT id FROM campaigns WHERE name = 'Spring2026'),
 'reconstruction', 'reconstruction_dev', 'in_progress', 'high',
 '~500 GB', '1M events', 'Spring2026', 'key4hep-2025-11-28', 'CLD, IDEA',
 'EDM4hep', '2026-05-01',
 'Must include simhit collections — do not strip truth info.',
 'fcc-ee,tracker,ttbar,benchmarking',
 'approved', 'approved',
 '2026-02-20 10:00:00', '2026-04-08 12:00:00'),

-- 15: approved, campaign-assigned
('$H \to \gamma\gamma$ at FCC-ee 240 GeV',
 'Rare Higgs decay to two photons. High-granularity ECAL response is critical; need truth-level photon isolation flags.',
 'Li Chen', 'lchen', 'lchen@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Higgs physics'),
 (SELECT id FROM campaigns WHERE name = 'Summer2026'),
 'simulation', 'physics_analysis', 'approved', 'high',
 '~1 TB', '5M signal events', 'Summer2026', 'key4hep-2025-11-28', 'IDEA',
 'EDM4hep', '2026-06-01',
 'Full calorimeter segmentation. Store all shower sub-clusters.',
 'fcc-ee,higgs,diphoton',
 'approved', 'approved',
 '2026-02-28 14:00:00', '2026-04-03 10:00:00'),

-- 16: cancelled
('Vector boson scattering: $W^+W^-jj$ at FCC-hh',
 'EW production of same-sign WW via VBS topology for Higgs-gauge coupling unitarisation studies.',
 'Hans Mueller', 'hmueller', 'hmueller@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Higgs physics'), NULL,
 'generation', 'physics_analysis', 'cancelled', 'medium',
 '~4 TB', '50M events', '', '', '',
 'HepMC3', '2026-08-01',
 'Cancelled: overlap identified with EW WG production request. Merging into that request.',
 'fcc-hh,vbs,ww,higgs',
 '', '',
 '2026-01-25 10:00:00', '2026-03-05 09:00:00'),

-- 17: pending, delphes stage with three detectors — the headline multi-detector example
('CLD vs. IDEA vs. ALLEGRO fast-sim benchmarking sample',
 'Common hard-scatter sample ($t\bar{t}$, VH, dijets) run through fast simulation for all three detector concepts, for the detector comparison paper. Identical generator-level input required for all three — see #5 for the shared generator setup.',
 'Sofia Fernandez', 'sfernandez', 'sfernandez@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'Geometry, Simulation'), NULL,
 'delphes', 'benchmarking', 'pending', 'medium',
 '~2 TB', '1M events per process per detector', 'Summer2026', 'key4hep-2025-11-28', 'CLD, IDEA, ALLEGRO',
 'EDM4hep', '2026-07-15',
 'Store reco objects only.',
 'fcc-ee,cld,idea,allegro,benchmarking,detector-comparison',
 '', '',
 '2026-03-25 09:00:00', '2026-03-25 09:00:00'),

-- 18: pending, generation
('High-mass $Z'' \to \ell\ell$ signal at FCC-hh',
 'Sequential SM $Z''$ signal samples from 1 to 40 TeV for dilepton resonance search sensitivity estimates.',
 'Thomas Dupont', 'tdupont', 'tdupont@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'BSM physics'), NULL,
 'generation', 'physics_analysis', 'pending', 'high',
 '~500 GB', '200k events per mass point (20 points)', '', '', '',
 'HepMC3', '2026-09-15',
 'Pythia8 SSM $Z''$ model. Detector-level smearing applied downstream via a Delphes FCC-hh card for a fast estimate.',
 'fcc-hh,bsm,zprime,dilepton',
 '', '',
 '2026-04-11 10:00:00', '2026-04-11 10:00:00'),

-- 19: pending, generation, relation partner for #8
('Graviton KK → ZZ at FCC-hh: signal',
 'Kaluza-Klein graviton signal samples for spin-2 resonance search in ZZ → llll final state. Shares the same production pipeline as the gluino grid scan (#8).',
 'Li Chen', 'lchen', 'lchen@cern.ch',
 (SELECT id FROM coordinator_groups WHERE name = 'BSM physics'), NULL,
 'generation', 'physics_analysis', 'pending', 'medium',
 '', '500k events per mass point (1–10 TeV, 10 points)', '', '', '',
 'HepMC3', '2026-09-01',
 'MadGraph5 RS1 model. Include both gg and qq̄ initial states.',
 'fcc-hh,bsm,graviton,kk,zz',
 '', '',
 '2026-03-30 10:00:00', '2026-03-30 10:00:00'),

-- 20: draft, "other" dataset type and use case
('Exploratory: timing-layer hit sample for pile-up rejection studies',
 'Early exploratory sample to scope out whether a dedicated timing-layer request is worth pursuing. Not yet a formal production request.',
 'Elena Patel', 'epatel', 'epatel@cern.ch',
 NULL, NULL,
 'other', 'other', 'draft', 'low',
 '', '', '', '', '',
 '', '',
 'Still deciding on the right generator-level setup.',
 'fcc-hh,timing,exploratory',
 '', '',
 '2026-04-12 09:00:00', '2026-04-12 09:00:00');

-- ── Production IDs ──────────────────────────────────────────────────────────

INSERT INTO production_ids (request_id, label, production_id, created_by, created_at) VALUES
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Higgs → bb̄%'), 'Generation', 128501, (SELECT id FROM users WHERE username = 'epatel'), '2026-02-01 09:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Higgs → bb̄%'), 'Full Sim',   128502, (SELECT id FROM users WHERE username = 'epatel'), '2026-02-20 11:30:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'ML training sample%'), 'Reconstruction', 129044, (SELECT id FROM users WHERE username = 'rkim'), '2026-02-05 10:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Track reconstruction benchmarking%'), 'CLD',  129301, (SELECT id FROM users WHERE username = 'rkim'), '2026-03-01 08:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Track reconstruction benchmarking%'), 'IDEA', 129302, (SELECT id FROM users WHERE username = 'rkim'), '2026-03-01 08:05:00');

-- ── Activity: comments, internal notes, mentions ────────────────────────────

INSERT INTO request_activity (request_id, user_id, type, body, created_at) VALUES
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Higgs → bb̄%'),
   (SELECT id FROM users WHERE username = 'epatel'), 'comment',
   '## Status update

Generation finished, full sim is **~60% done** on the grid. Current throughput:

| Stage       | Events done | ETA        |
|-------------|------------:|------------|
| Generation  | 10M         | done       |
| Full sim    | 6M          | ~2 weeks   |

See the [Delphes card used for the cross-check](https://github.com/HEP-FCC/FCC-config) and $\sigma(e^+e^- \to ZH) \approx 200$ fb for reference.',
   '2026-03-01 10:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Higgs → bb̄%'),
   (SELECT id FROM users WHERE username = 'epatel'), 'internal_note',
   'Grid site CNAF has been flaky this week — re-submitted the failed subset of jobs, should not affect the public ETA above.',
   '2026-03-02 09:15:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Higgs → bb̄%'),
   (SELECT id FROM users WHERE username = 'akowalski'), 'comment',
   'Thanks for the update! Linking the single-top cross-check request here for visibility: #13',
   '2026-03-10 14:22:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'ML training sample%'),
   (SELECT id FROM users WHERE username = 'mrossi'), 'comment',
   'Class balance in the first 20M-jet batch:

- b-jets: 25%
- c-jets: 25%
- light: 25%
- gluon: 25%

Looks good — proceeding with the full 200M.',
   '2026-02-10 09:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'PID benchmark%'),
   (SELECT id FROM users WHERE username = 'rkim'), 'internal_note',
   'DIRAC production 127900 failed at the kaon 50 GeV point — `pid-rich-sim` app segfaulted on the Manchester site. Marked as **failed** rather than resubmitting since the campaign has already closed.',
   '2026-10-18 08:00:00');

-- ── Relations between requests ──────────────────────────────────────────────

INSERT INTO request_relations (from_id, to_id, type, created_by, created_at) VALUES
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Single-top associated production%'),
   (SELECT id FROM dataset_requests WHERE title LIKE 'Higgs → bb̄%'),
   'related', (SELECT id FROM users WHERE username = 'akowalski'), '2026-04-02 09:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Single-top associated production%'),
   (SELECT id FROM dataset_requests WHERE title LIKE 'Tracker alignment sample%'),
   'depends_on', (SELECT id FROM users WHERE username = 'akowalski'), '2026-04-02 09:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'Graviton KK%'),
   (SELECT id FROM dataset_requests WHERE title LIKE 'Gluino pair production%'),
   'variant', (SELECT id FROM users WHERE username = 'lchen'), '2026-03-30 10:00:00'),
  ((SELECT id FROM dataset_requests WHERE title LIKE 'CLD vs. IDEA vs. ALLEGRO%'),
   (SELECT id FROM dataset_requests WHERE title LIKE 'Tracker alignment sample%'),
   'depends_on', (SELECT id FROM users WHERE username = 'sfernandez'), '2026-03-25 09:00:00');

-- ── Generator card (plain-text Pythia8 config, stored as a small blob) ──────

INSERT INTO generator_cards (request_id, filename, size, content, uploaded_by, created_at)
SELECT
  (SELECT id FROM dataset_requests WHERE title LIKE 'ECAL single-particle response%'),
  'egun.cmnd',
  LENGTH(c), c,
  (SELECT id FROM users WHERE username = 'akowalski'),
  '2025-08-16 09:30:00'
FROM (SELECT '! Single-electron gun, log-spaced energy points
Beams:idA = 11
Beams:idB = -11
PartonLevel:ISR = off
PartonLevel:FSR = off
ParticleGun:enable = on
ParticleGun:id = 11
ParticleGun:eMin = 1.0
ParticleGun:eMax = 100.0
ParticleGun:etaMin = -3.0
ParticleGun:etaMax = 3.0
' AS c);
