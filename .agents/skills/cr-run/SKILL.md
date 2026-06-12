---
name: cr-run
description: >-
  Use this skill when the user explicitly asks to execute/implement a CR plan produced by cr-plan — for example "cr-run this", "run the CR plan", "implement the cr-plan", "let's build the CR". Works a plan at docs/plans/cr-<cr-name>.md one subtask at a time, in order: build the subtask, verify it by running its tests, then commit that subtask on its own before moving on. Tests are the only gate — no human confirmation between subtasks. The CR lands as one verified commit per subtask. Requires an existing plan; if none exists, points the user at cr-plan first.
---

# CR Run

A skill for executing a CR plan produced by **cr-plan**. It takes the plan at `docs/plans/cr-<cr-name>.md` and works through the subtasks **in order**, one at a time. Each subtask is built, then verified by its tests before it counts as done, then committed on its own. The CR ends up as a sequence of verified commits — one per subtask.

The completion protocol is non-negotiable and comes straight from the plan's **Definition of done**:

> A subtask is done when **its tests pass.** Then it is committed on its own, before the next subtask begins.

The tests are the only gate — there is no human confirmation step. That puts the weight on the tests actually covering the plan's per-subtask test scenarios (happy path + the edge/failure cases it lists): a trivial test that passes without exercising the deliverable does **not** count as done. No batching subtasks into one commit, no skipping ahead past a subtask whose tests aren't green.

This skill owns the *how* — writing the code, the tests, running them, committing. It does **not** redefine the *what* or *why*; those are fixed by the plan. If the plan turns out to be wrong, that's a signal to go back to cr-plan, not to improvise a different deliverable here.

## When to use this skill

Use this skill when the user explicitly asks to execute or implement a CR plan. Typical openings:

- "cr-run the auth migration."
- "Let's implement the cr-plan."
- "Run the CR plan and commit as you go."

If there's **no plan yet**, this skill is not the right fit — point the user at **cr-plan** first. Don't reconstruct a plan from memory and start coding; the plan (with its testable deliverables and per-subtask test scenarios) is the contract this skill executes against.

## How this fits with cr-plan

This skill is the fourth link in the chain: **anchor-def** (the problem) → **feature-def** (goals, requirements, acceptance criteria) → **cr-plan** (executable, testable plan) → **cr-run** (execute it).

- The plan is the input. Read it in full before touching code: the why, the overall deliverable, every subtask's deliverable and test scenarios, the definition of done, and any open questions/risks.
- cr-run keeps the plan's **Status** and per-subtask **Status** fields up to date as it goes, so the artifact always reflects reality.

## Core principles

1. **Done means tests green, then committed.** A subtask is done when its tests pass — no human confirmation. Then commit. This is the heart of the skill; everything else serves it.
2. **One subtask at a time, in order.** Finish, verify, and commit a subtask before starting the next. Respect declared dependencies. No parallel half-finished work.
3. **One commit per verified subtask.** Focused commit, message describing the delivered outcome. Don't batch, don't commit unverified work.
4. **Execute the plan; don't rewrite it.** The what and why are fixed. If reality contradicts the plan, stop and send the user back to cr-plan rather than silently changing the deliverable.
5. **Tests are the contract — and the only gate.** Implement against the subtask's test scenarios from the plan, and make the tests genuinely cover them (happy path + edge/failure). Since nothing else confirms the work, a weak or trivial test that passes is a failure of the protocol. If a scenario can't be satisfied as written, surface it — don't quietly weaken the test to make it pass.

## The flow

1. **Load the plan** — locate and read `docs/plans/cr-<cr-name>.md`; confirm it's ready.
2. **Pick the starting point** — find the first subtask not yet done; respect dependencies.
3. **Run the subtask loop** — for each subtask: build → write & run tests → commit when green → update status.
4. **Close out the CR** — when every subtask is done and committed, confirm the overall deliverable holds.

## Step 1 — Load the plan

- Locate the plan at `docs/plans/cr-<cr-name>.md` (ask the user which CR if ambiguous).
- If no plan exists, stop and point the user at **cr-plan**. Do not start coding without one.
- Read the whole plan. Reflect back to the user, briefly: the why, the overall deliverable, and the list of subtasks with their current status.
- If the plan's **Status** is not "Ready for implementation", flag it — the plan may not be verified yet. Offer to go back to cr-plan to finish verifying before executing.

## Step 2 — Pick the starting point

- Identify the first subtask whose **Status** is "Not started" (or the one the user names).
- Check its **Depends on** field — don't start a subtask whose dependencies aren't done.
- Tell the user where you're starting and the sequence you'll work, then proceed. This is informational, not a gate — you don't wait for sign-off between subtasks.

## Step 3 — Run the subtask loop

For each subtask, in order — no pause for sign-off between them:

1. **Build it.** Implement the deliverable. Stay within this subtask's scope — resist pulling in later subtasks' work.
2. **Write the tests** for this subtask's test scenarios from the plan (happy path + the edge/failure cases it lists). They must genuinely exercise the deliverable — the tests are the only gate, so a trivial pass doesn't count. Prefer unit tests; use integration/e2e only where the scenario requires it.
3. **Run the tests.** Run this subtask's tests *and* the existing suite. They must be green. If they fail, fix and re-run; do not proceed while red.
4. **Commit.** Once green, commit this subtask on its own with a focused message describing the delivered outcome.
5. **Update status.** Set the subtask's **Status** in the plan to `Done — committed <sha>` and save the plan.
6. **Move to the next subtask.**

Keep the user informed as you go (what landed, tests green, committed), but don't block on a reply. Commit only on green — never commit a subtask whose tests are red or missing.

If implementing a subtask reveals the plan is wrong (a deliverable doesn't make sense, a dependency was missed, a scenario is untestable), **stop**. Don't improvise a different deliverable. Surface it to the user and recommend a quick pass back through cr-plan to fix the plan, then resume.

## Step 4 — Close out the CR

When every subtask is `Done — committed`:

- Run the **full test suite** once more; confirm green.
- Re-read the plan's **What this CR delivers** and confirm the union of the committed subtasks actually delivers it. If there's a gap, it's a new (or amended) subtask — back to Step 3, not a silent fix.
- Update the plan's top-level **Status** to "Implemented".
- Summarize for the user: the subtasks delivered and their commits, and confirm the CR's overall deliverable is met.

## Handling edge cases

- **No plan exists.** Stop and point the user at cr-plan. Don't reconstruct a plan and start coding.
- **Plan isn't verified ("Ready for implementation").** Flag it; offer to finish verifying via cr-plan before executing.
- **Tests pass but look trivial or don't really exercise the deliverable.** Green is the only gate, so weak tests are the main risk now. Make the tests cover the plan's scenarios (happy + edge). If the plan's scenarios themselves are too thin to trust, surface it and route back to cr-plan.
- **A subtask's tests won't pass.** Fix the implementation. If the *scenario itself* is wrong or untestable, surface it and go back to cr-plan — don't weaken the test just to get green.
- **The plan is wrong once you start building.** Stop and route back to cr-plan. Don't silently change the deliverable or scope.
- **The user asks to skip a subtask or reorder.** Honor an explicit request, but check dependencies first and call out any deliverable that would be left untested or unmet.
- **The user asks to batch commits or skip committing.** Surface the trade-off (one verified commit per subtask is the protocol) and honor an explicit override — but don't silently drop the commit step or commit unverified work.
- **Mid-execution resume.** Read the per-subtask **Status** fields to find where you left off; don't redo committed subtasks.

## Tone

Act as a disciplined implementer who treats green tests as the definition of done — and therefore writes tests worth trusting. Be matter-of-fact about what passed, what didn't, and what's left. Show the test results as you go. The user should be able to see which subtask is in flight without being asked to approve each one.
