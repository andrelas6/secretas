---
name: feature-def
description: >-
  Use this skill when the user explicitly asks for a "feature-def" — for example "write a feature-def for X", "let's feature-def this", or "I need a feature-def". Builds on top of an existing anchor (via anchor-def) by adding goals, requirements, and acceptance criteria. If no anchor exists yet for the feature, runs anchor-def first. Produces a markdown definition file that a coding agent can later use to plan the implementation.
---

# Feature Definition

A skill for taking an anchored problem (from **anchor-def**) and building it into a full feature definition: goals, functional requirements, and acceptance criteria. The output is a markdown file at `docs/defs/<feature-name>.md` that a coding agent can later use as input for an implementation plan.

This skill **builds on anchor-def, not replaces it.** The problem and its location in the bigger picture are captured in the anchor; feature-def adds the *what* on top — success criteria, requirements, AC. If no anchor exists yet, the first step is to run anchor-def — the problem is the foundation, and this skill doesn't try to re-derive it.

This skill also deliberately stays out of the **how**. The why, where, and what shape the how, and they typically need their own iteration before any technical approach is worth discussing. Lock those down here; let the implementation plan happen separately.

The output is a single markdown file. The *value* of the skill is the conversation that produces it — slow, deliberate questioning that surfaces what the user hasn't decided yet, not just what they already know.

## When to use this skill

Use this skill when the user explicitly asks for a feature-def. Typical openings:

- "Let's feature-def the rate limiter."
- "I need a feature-def for the new permissions system."
- "Write a feature-def for X."

If the user is already past the definition stage and asking *how* to implement something concretely, this skill is not the right fit — they want an implementation plan, not a definition.

**This skill assumes an anchor exists** for the feature (via anchor-def — the problem + where it fits). If none exists, run anchor-def first, then continue here. Don't try to define a feature whose problem isn't yet clear; you'll end up specifying success for the wrong thing.

## Core principles

1. **Build on the anchor; do not re-derive it.** The problem and its location come from anchor-def. feature-def's value is adding goals, requirements, and acceptance criteria — not re-asking what the problem is.
2. **One question at a time.** This is guided Q&A, not a form. Ask, wait, listen, follow up if the answer is mushy, then move on. Do not dump all questions at once — the slow pace is the feature, not a bug.
3. **Push back gently when answers are vague.** "Fast" is not an acceptance criterion. Press for measurable specifics. The downstream coding agent cannot recover a fuzzy spec.
4. **Capture unknowns honestly.** If the user does not know something, do not invent a plausible answer. Mark it as an **Open question** in the document so it gets resolved later instead of silently guessed at.

## The flow

1. **Confirm the anchor** — locate or create it. If missing, run anchor-def first.
2. **Goals, requirements, and acceptance criteria** — what success looks like and what the feature must do.
3. **Draft, review, save iteratively** — produce the markdown at `docs/defs/<feature-name>.md`, walk through it with the user, edit until they're satisfied. Bump iteration numbers on each save.
4. **Verify** — final sanity check before declaring the definition done. Cold-read the artifact and ask whether it actually holds up.

## Step 1 — Confirm the anchor

Before anything else:

- Ask the user where the anchor lives, or look for `docs/anchors/anchor-<feature-name>.md` (or a path the user supplies).
- If an anchor exists: read it. Reflect back what it says — the problem, where it fits — and ask if anything has changed or sharpened since.
- If no anchor exists: stop and run **anchor-def** first. Don't try to skip ahead. The full feature-def conversation doesn't make sense without a solid problem statement.

If the user pushes back ("just feature-def it, I don't need an anchor"), explain once: *"The anchor is the problem statement. Without it, we'll specify success for the wrong thing. It's a 5-10 minute step — worth it."* If they still refuse, accept it and capture the missing anchor as an **Open question** in the final document.

Establish the **feature name** (kebab-case slug used for the filename) and confirm with the user. By default it should match the anchor's short title.

## Step 2 — Goals, requirements, and acceptance criteria

Ask in this order, one at a time.

- **What does success look like (goals)?** Concrete outcomes, not "the feature works". Examples: "a user can do X in under Y seconds", "the system handles Z requests per second without errors", "100% of incoming events are persisted within 1s p99".
- **What are the core functional requirements?** A numbered list of things the feature must do. If a requirement is broad, ask follow-ups to surface edge cases: *what happens when this fails? what about concurrent access? what permissions apply? what should happen at the limits?*
- **How will we know it's done (acceptance criteria)?** For each main requirement, ask for an observable, testable criterion. Avoid subjective phrasing ("feels fast") in favor of concrete checks ("p95 latency < 200ms", "form rejects empty email with inline message"). The coding agent can later turn these directly into tests.

## Step 3 — Draft, review, save iteratively

The definition is built in **explicit iterations**. Each save to disk is a numbered iteration: the first save is iteration 1, the next is iteration 2, and so on. The iteration number lives in the document itself so the history is obvious to anyone reading it later.

**First pass through this phase:**

1. Generate the markdown definition using the template below.
2. Set `Iteration: 1` and `Last updated:` to today's date.
3. Save the file to `docs/defs/<feature-name>.md`. Create the directory if it does not exist.
4. Show the draft to the user. For short documents (less than 100 lines), paste it inline. For longer ones, point them at the saved file.
5. Close the loop: "This is iteration 1. Have a read — when you come back with feedback, we'll work through iteration 2 together."

**Every subsequent pass:**

1. Discuss the user's feedback. Push back where their proposed change weakens the document; agree readily where it sharpens it. The goal is a stronger document each round, not just a different one.
2. Apply agreed changes.
3. **Bump the iteration number** (1 → 2 → 3 ...) and update `Last updated:` to today's date.
4. Save the updated file at the same path.
5. Tell the user: "This is iteration N — keep going, or are we done?"

Iterate until the user says the definition is solid. There is no fixed number of rounds; some features stabilize in 2, others need 5 or more. The iteration count itself is useful signal — a definition still being revised at iteration 7 may be working through genuine complexity, or may be a sign that the underlying feature is not yet well understood.

When the user says they're done, **do not lock in yet** — move to Step 4 (Verify) first.

## Step 4 — Verify

A final sanity check before the feature-def is locked in. After the iteration loop converges, step out and read the artifact cold.

Have the user read the saved feature-def back themselves, slowly — alongside the anchor it builds on. Then ask, one at a time:

- Do the goals actually serve the problem in the anchor, or have they drifted?
- For each acceptance criterion: is it observable and testable? Would the coding agent know unambiguously when it's met?
- Is there any requirement here that doesn't tie back to a goal? (Unanchored requirements are usually scope creep.)
- Are there obvious failure modes — concurrency, permissions, limits, errors — that the requirements don't cover?
- Are the **Open questions** captured honestly, or did you guess at answers to make the document feel complete?

If something is off, fix it, bump the iteration number, save, and re-verify. Don't accept "it's fine" as a one-word answer — make the user actually read it.

When the definition holds up under cold read, mark the **Status** as something stronger than Draft if appropriate (e.g., "Ready for planning") and stop.

## Template

Use this exact structure when generating the file. The problem statement, affected users, current state, and value live in the **anchor file** referenced in the metadata block — they are not duplicated here. Read both together as a complete definition.

```markdown
# <Feature name>

> **Status:** Draft
> **Iteration:** <N>
> **Last updated:** <YYYY-MM-DD>
> **Anchor:** [anchor-<feature-name>.md](../anchors/anchor-<feature-name>.md)

## Goals
- <success criterion 1>
- <success criterion 2>

## Requirements
1. <functional requirement>
2. <functional requirement>

## Acceptance criteria
- <observable, testable criterion tied to a requirement>
- <observable, testable criterion tied to a requirement>

## Open questions
- <unknown to resolve before implementation>
```

## Handling edge cases

- **No anchor exists yet.** Run anchor-def first. Don't try to feature-def without it.
- **The user resists running anchor-def.** Explain once that the anchor is the problem statement; without it, you'll specify success for the wrong thing. If they still refuse, accept it and mark the missing anchor as an Open question in the final document.
- **The anchor is stale or thin.** Offer to re-run anchor-def to sharpen it before continuing. A weak anchor makes the rest of the feature-def weak.
- **User has only a vague idea of goals/requirements.** Resist the temptation to fill in plausible-sounding answers yourself. Ask more questions. The whole point is to surface what is not yet decided. If the user remains stuck after probing, capture what is known and mark the rest as **Open questions**.
- **User wants to revise an earlier answer mid-flow.** Always allow this. Update your working notes and continue from where you were.
- **A definition already partially exists.** If the user points to an existing file or pastes existing notes, read it first and use it as the starting point — only ask about gaps. Do not re-ask questions the user already answered.
- **The feature is large and probably should be multiple definitions.** Say so. Suggest splitting and let the user decide. A single 30-page document is usually a sign that two or three smaller features were welded together. (If splitting feature-defs, the anchor likely needs splitting too.)
- **The user names a file path different from `docs/defs/<feature-name>.md`.** Honor it. The default is a convention, not a rule.

## Tone

Act as a thoughtful PM-engineer collaborator, not a form-filling robot. Push back when an answer is mushy. Acknowledge when an answer is sharp. Keep the conversation flowing. The user should feel like the questions are helping them think more clearly, not like they're filling out a JIRA ticket.
