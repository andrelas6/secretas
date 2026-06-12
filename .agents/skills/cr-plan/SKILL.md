---
name: cr-plan
description: >-
  Use this skill when the user explicitly asks to plan a CR (change request) — for example "cr-plan this", "plan this CR", "break this change into a plan". Turns a change request into an executable plan for a coding agent: the why (what it solves), ordered subtasks each with a concrete deliverable, and test scenarios for every deliverable. Every deliverable must be independently testable, and the CR as a whole must deliver something testable. Reuses the "why" from an existing anchor-def/feature-def if one exists; otherwise asks for it directly. Produces a markdown plan file.
---

# CR Plan

A skill for turning a change request into a plan a coding agent can execute. The output is a markdown file at `docs/plans/cr-<cr-name>.md` containing three things the implementing agent needs: **why** the CR exists (what it solves), an ordered list of **subtasks each with a concrete deliverable**, and **test scenarios for every deliverable**.

Two hard rules shape every plan this skill produces:

1. **Every deliverable must be independently testable** — preferably with unit tests. A subtask whose result can't be verified isn't a deliverable; it's a hope.
2. **The CR as a whole must clearly deliver something testable.** If you can't name what the CR delivers and how you'd test it, the CR isn't ready to plan — go back to definition.

**This skill only plans — it does not implement.** Execution is a separate skill (**cr-run**), which works the plan one subtask at a time, verifying and committing each. To make that hand-off clean, the plan records the **definition of done** every subtask must meet: it isn't done until its tests pass, and each verified subtask is then committed on its own. The tests are the only gate — no human sign-off — so the plan's per-subtask test scenarios have to be strong enough to stand on their own. cr-plan writes that protocol into the artifact; cr-run carries it out.

This skill stays out of *deep* implementation detail. It defines *what* each subtask delivers and *how you'll know it works*, not the line-by-line *how*. The agent that executes the plan owns the how; this skill makes sure the agent is aiming at testable, well-ordered targets.

The output is a single markdown file. The *value* of the skill is the conversation that produces it — deliberate questioning that forces the work into testable pieces before any code is written.

## When to use this skill

Use this skill when the user explicitly asks to plan a CR. Typical openings:

- "cr-plan the auth migration."
- "Let's plan this CR."
- "Break this change into a plan I can hand to an agent."

If the user is still working out *what* the change is or *why* it matters, that's definition work — point them at **feature-def** (or **anchor-def**) first. This skill plans a change whose intent is already clear; it doesn't discover the intent.

## How this fits with anchor-def and feature-def

This skill is the third link in the chain: **anchor-def** (the problem) → **feature-def** (goals, requirements, acceptance criteria) → **cr-plan** (executable, testable plan) → **cr-run** (execute the plan, verifying and committing each subtask).

- The **why** comes from an existing anchor or feature-def when one exists. Look for `docs/defs/<name>.md` or `docs/anchors/anchor-<name>.md`, read it, and pull the why and acceptance criteria from there — don't re-derive them.
- If **no** def exists, ask the user for the why directly (Step 1). A CR with no stated why is not plannable; the whole point of the plan is to keep the work pointed at what it solves.

## Core principles

1. **Testable or it doesn't ship.** Every deliverable carries its own test scenarios. The CR as a whole delivers something testable. No untestable subtasks, no untestable CR.
2. **Smallest testable slice is the target.** The unit of work is the smallest change that still delivers something you can write a meaningful test for. Drive the split down to that floor — go finer than feels necessary. Don't stop at "one testable deliverable" if it can be peeled into two thinner testable slices.
3. **Subtasks are deliverables, not steps.** "Refactor the parser" is a step. "Parser accepts the new field and rejects malformed input" is a deliverable with a clear test. Push every subtask toward a verifiable outcome.
4. **Order for verifiability.** Sequence subtasks so each one can be tested as it lands, ideally building on the last. A plan where nothing is testable until the end is a planning failure.
5. **One question at a time.** This is guided Q&A, not a form. Ask, wait, listen, follow up if the answer is mushy, then move on. The slow pace is the feature, not a bug.
6. **Don't invent the why.** If the why isn't in a def and the user can't state it, stop. Capture it as an Open question rather than guessing.
7. **Record the definition of done; don't execute it.** Every plan states that a subtask is done when its tests pass, then is committed on its own. The tests are the only gate, so the test scenarios must be strong enough to carry that weight. cr-plan writes this into the artifact so cr-run can carry it out — but cr-plan itself never implements, tests, or commits the work.

## The flow

1. **Establish the why** — reuse from an existing def, or ask the user. What does this CR solve, and what for?
2. **Name the CR's overall deliverable** — the one testable thing this CR ships.
3. **Break into subtasks with deliverables** — ordered, each independently testable.
4. **Write test scenarios per deliverable** — concrete, observable checks.
5. **Draft, review, save iteratively** — produce the markdown, walk through it, edit until solid. Bump iteration numbers.
6. **Verify** — cold-read the plan and check every deliverable is genuinely testable.

Execution is out of scope — once the plan is verified, hand it to **cr-run**.

## Step 1 — Establish the why

First, look for an existing definition:

- Check `docs/defs/<cr-name>.md` and `docs/anchors/anchor-<cr-name>.md`, or ask the user where the def lives.
- **If a def exists:** read it. Reflect back the why and the acceptance criteria, and ask if anything has sharpened since. Use these as the foundation — the acceptance criteria often become the CR's overall test target.
- **If no def exists:** ask the user directly, one question at a time:
  - *Why are you doing this CR — what problem does it solve?*
  - *What is it for — what becomes possible or better once it lands?*

Push back if the why is vague ("clean things up", "make it better"). A why you can't test against produces deliverables you can't test. If the user genuinely can't state it, capture it as an **Open question** and consider suggesting they run anchor-def first.

Establish the **CR name** (kebab-case slug for the filename) and confirm it.

## Step 2 — Name the CR's overall deliverable

Before breaking anything down, force one sentence: *what does this CR deliver, and how would you know it works?*

This is the testability gate for the whole CR. If the answer is observable and testable, proceed. If it isn't — if the CR's outcome can't be verified — the CR isn't ready; the user needs to sharpen the definition first (feature-def), not plan it here.

## Step 3 — Break into subtasks with deliverables

Drive the split down to the **smallest unit that still delivers something independently testable** — the smallest change you could make and write a meaningful test for. Go finer than feels necessary: if you can peel a thinner slice that still has something to test, peel it. (A hello-world endpoint is two slices, not one: "returns 200 with an empty body", then "returns the JSON payload".)

Two decomposition patterns get you to that floor — pick whichever fits the work:

**Pattern 1 — Skeleton-first (progressive refinement).** Stand up the simplest end-to-end version that works, then each subsequent subtask *deepens* it. Every subtask leaves a working (if trivial) product. Best when there's a behavior you can stand up thin and fatten.
- *Example (REST write endpoint):* controller returns 201 with mocked data → parse and validate the request body (400 on bad input) → replace the mock with the real repository call.
- Reliable sub-technique: start on **stubs/mocks**, then make each subtask **replace a stub with the real thing** — each replacement is its own testable slice.

**Pattern 2 — Stage-by-stage (build-to-completion).** Decompose along the stages of a transformation/pipeline. Each stage is independently testable, but the CR only delivers its end goal once the final stage lands. Best when the work is a pipeline.
- *Example (static-site generator):* load markdown files → array of Post objects → extract metadata (slug, title) → extract body content → convert content to HTML → write HTML to the correct destination path. Each stage gets its own test; the pipeline is incomplete until the last.

Then:

- **Don't split *below* testability.** A bare function signature or scaffolding with nothing to assert is not a slice — fold it into the first subtask that actually exercises it.
- **Order** subtasks so each is testable as it lands, building toward the overall deliverable from Step 2 and respecting dependencies.
- For each subtask, capture: a short title, the **deliverable** (the concrete, testable outcome), and any dependency on earlier subtasks.
- **Note which pattern you used** — it goes in the plan's **Approach** line, so the reader knows the shape (and, for stage-by-stage, that intermediate subtasks don't yet deliver the end goal).

If a subtask's deliverable can't be made independently testable no matter how it's reshaped, say so plainly and capture it as a risk or Open question — don't paper over it.

## Step 4 — Write test scenarios per deliverable

For every deliverable, write the **test scenarios** that prove it works. One at a time, per subtask:

- Cover the happy path *and* the meaningful failure/edge cases (bad input, limits, concurrency, permissions — whatever applies).
- Make each scenario observable: given/when/then, or a plain "test that X produces Y". Avoid "works correctly" — name the input and the expected output.
- Prefer scenarios that map to unit tests. Note where a scenario needs integration/e2e instead, so the agent knows.

These scenarios are the contract the implementing agent codes against. They should be concrete enough to turn directly into tests.

## Step 5 — Draft, review, save iteratively

The plan is built in **explicit iterations**. Each save to disk is a numbered iteration.

**First pass:**

1. Generate the markdown plan using the template below.
2. Set `Iteration: 1` and `Last updated:` to today's date.
3. Save to `docs/plans/cr-<cr-name>.md`. Create the directory if it doesn't exist.
4. Show the draft (inline if under ~100 lines, otherwise point at the file).
5. Close the loop: "This is iteration 1 — read it through and we'll work iteration 2 from your feedback."

**Every subsequent pass:**

1. Discuss feedback. Push back where a change weakens testability; agree where it sharpens the plan.
2. Apply agreed changes.
3. Bump the iteration number and update `Last updated:`.
4. Save at the same path.
5. Tell the user: "This is iteration N — keep going, or are we done?"

Iterate until the user says the plan is solid. Then move to Step 6 before declaring done.

## Step 6 — Verify

A final cold read before the plan is handed off. Have the user read the saved plan back slowly — alongside the def it builds on, if any. Then ask, one at a time:

- Does the **why** still hold, and does every subtask actually serve it?
- For **each deliverable**: is it independently testable? Could the agent verify it without finishing the rest of the CR?
- **Granularity:** could any subtask be cut into two independently-testable slices? If yes, cut it — the smallest testable slice is the target.
- For **each test scenario**: is it concrete and observable? Could the agent turn it into a test without guessing the expected result? Is it strong enough to be the *only* gate (no human will confirm)?
- Does the CR's **overall deliverable** (Step 2) end up genuinely tested by the union of the subtask scenarios?
- Is the ordering right — can each subtask be tested as it lands?
- Does the decomposition follow one pattern (skeleton-first or stage-by-stage) coherently, rather than a confused mix?
- Are **Open questions / risks** captured honestly, or did you guess to make the plan feel complete?

If something is off, fix it, bump the iteration, save, and re-verify. Don't accept "it's fine" — make the user actually read it. When the plan holds up, mark **Status** as "Ready for implementation" and stop. Then point the user at **cr-run** to execute it.

## Template

Use this exact structure when generating the file. If the why and acceptance criteria come from a def, reference it in the metadata rather than duplicating the full problem statement.

```markdown
# CR plan: <cr-name>

> **Status:** Draft
> **Iteration:** <N>
> **Last updated:** <YYYY-MM-DD>
> **Def:** [<name>.md](../defs/<name>.md)  <!-- omit if no def exists -->

## Why this CR
<What problem it solves and what it's for. Pulled from the def if one exists, or stated directly.>

## What this CR delivers
<One or two sentences: the single testable outcome this CR ships, and how you'd know it works.>

## Approach
<Skeleton-first (progressive refinement) or Stage-by-stage (build-to-completion), plus one line on the shape of the split. For stage-by-stage, note that the end goal is only delivered by the final subtask.>

## Definition of done (per subtask)
Each subtask is done when **its tests pass** (and the tests genuinely cover this subtask's test scenarios). Each verified subtask is committed on its own before the next begins. There is no human confirmation step — the tests are the only gate.

## Subtasks

### 1. <subtask title>
- **Deliverable:** <concrete, independently testable outcome>
- **Depends on:** <earlier subtask, or "none">
- **Test scenarios:**
  - <given / when / then — happy path>
  - <edge / failure case>
- **Status:** Not started  <!-- → Done — committed <sha> -->

### 2. <subtask title>
- **Deliverable:** <concrete, independently testable outcome>
- **Depends on:** <earlier subtask, or "none">
- **Test scenarios:**
  - <scenario>
  - <scenario>
- **Status:** Not started  <!-- → Done — committed <sha> -->

## Open questions / risks
- <unknown or untestable area to resolve before/while implementing>
```

## Handling edge cases

- **No def exists.** Ask for the why directly (Step 1). If the user can't state it, suggest running anchor-def first and capture the gap as an Open question — don't invent a why.
- **The CR's overall outcome isn't testable.** Stop. The CR isn't ready to plan; the user needs to sharpen the definition (feature-def), not the plan. Say so plainly.
- **A subtask can't be made independently testable.** Try reshaping or reordering first. If it still can't, capture it as a risk/Open question rather than pretending it's testable.
- **The CR is really several CRs.** If the deliverables don't cohere into one testable change, say so and suggest splitting into multiple CRs (one plan file each). Let the user decide.
- **The user wants to skip test scenarios "for now".** Hold the line gently: the test scenarios are the point — they're the contract the agent codes against. A plan without them is just a todo list.
- **The user asks cr-plan to start implementing.** That's cr-run's job. Finish and verify the plan here, then hand off. Don't write code, run tests, or commit from this skill.
- **The user revises an earlier answer mid-flow.** Always allow it. Update working notes and continue.
- **A plan already partially exists.** Read it first, use it as the starting point, only ask about gaps. Don't re-ask what's already answered.
- **The user names a different save path.** Honor it. The default is a convention, not a rule.

## Tone

Act as a thoughtful engineer-collaborator who refuses to let untestable work through. Push back when a deliverable is mushy or a test scenario is vague; acknowledge when one is sharp. The user should feel the plan getting more executable with each question — not like they're filling out a ticket.
