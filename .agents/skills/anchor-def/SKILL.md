---
name: anchor-def
description: >-
  A sensemaking skill the user invokes explicitly to anchor any problem and locate it in the bigger picture. Works for any domain — a coding feature, a bug, a process to optimize, a vacation to plan, an organizational question, a personal decision. Only triggers on explicit user intent: phrases like "anchor this", "anchor-def", "let me make sense of this first", "let me think this through". Do NOT trigger as a generic pre-action reflex; this is a deliberate sensemaking act. Produces a markdown doc with two sections: the problem (the anchor) and where it fits (immediate context + bigger picture).
---

# Anchor Definition

Sensemaking before action. Plant your feet, then look around.

A skill for forcing two questions to be answered clearly before committing to any course of action — *what is the problem, and where does it fit in the bigger picture* — and producing a tight markdown artifact that captures the answers. Domain-agnostic: works for a coding feature, a bug, a process to optimize, a vacation to plan, an organizational question, a personal decision.

This skill deliberately stays out of *how* to handle the problem. Anchoring is about understanding, not planning. The how can come after, separately, once the anchor is solid.

The output is a single markdown file. The *value* of the skill is the conversation that produces it — patient questioning that surfaces what the user hasn't actually nailed down yet, not just what they already think they know.

## When to use this skill

**This is sensemaking work, not a generic checklist.** It only runs when the user has explicitly chosen to make sense of something — not as a reflex any time action is about to be taken.

Trigger only when:

- The user explicitly invokes it ("anchor this", "anchor-def", "let me make sense of this first", "let me think this through") or similar language that signals intentional sensemaking.
- The user is visibly wrestling with **what** the problem actually is (not just *how* to handle it) and asks for a structured way to clarify it.

Do **not** trigger when:

- The user is in execution mode and hasn't asked to sensemake — even if it would technically help.
- The matter is trivial.
- A clear understanding already exists.
- The activity is pure exploration where the goal IS to discover, not to clarify.

If in doubt: ask the user "are you trying to make sense of this first, or are you ready to plan how to handle it?" Only run if they confirm the first.

## Core principles

1. **One question at a time.** This is guided Q&A, not a form. Ask, wait, listen, follow up if the answer is mushy, then move on. The slow pace is the feature, not a bug.
2. **The problem is the anchor.** You can't meaningfully zoom out from nowhere. Nail the problem first, then look around it. Never the other way around.
3. **No solution words in Step 1.** "We should…", "let's add…", "I'll do X" — push back every time. The problem statement comes before the solution. Don't paper over slipped solution language; it's a signal the user hasn't understood the problem yet.
4. **Patience is the whole point.** If the user is impatient and wants to move on before the anchor is solid, that's exactly the moment to slow down. A weak anchor makes everything downstream useless.

## The flow

1. **Define the problem (the anchor)** — one paragraph naming the gap as it exists today.
2. **Find where it fits** — immediate context first, then the bigger picture. Zoom out *from* the anchor.
3. **Draft, review, save** — produce the markdown file, walk through it with the user, edit until they're satisfied.
4. **Verify** — final sanity check before locking in. Cold-read the artifact and ask whether it actually holds up.

## Step 1 — Define the problem (the anchor)

Have the user write one paragraph naming the gap as it exists today. Probe until it's sharply defined. Useful probes:

- What's the gap? What exists today vs. what should exist?
- Who or what feels it? How do you know?
- What evidence makes this a problem and not just an opinion?
- If nothing changed, what would actually happen?

**Hard rule: no solution words in this section.** If the user writes "we should…", "let's add…", "I'll do X" — push back. Restate the rule. Ask them to try again.

The problem is defined well enough when:

- It's one paragraph, not a list.
- A stranger could read it and understand the gap without further explanation.
- It contains no proposed solution.
- It could be falsified — if X were true instead, this wouldn't be a problem.

Do not move to Step 2 until that bar is hit. Patience here is the whole point of the skill.

## Step 2 — Find where it fits

Two sub-prompts, in this order:

- **In its immediate context** — Where does this live? What does it touch? What surrounds it? What must stay stable nearby while addressing this?
- **In the bigger picture** — What larger goal or outcome does this serve? Why this, why now?

Order matters: immediate context first (one step out from the anchor), then bigger picture (a wider step out). Don't let the user write the bigger picture before the immediate context — without an anchor, the bigger picture is hand-waving.

## Step 3 — Draft, review, save

1. Generate the markdown using the template below.
2. Show it to the user. For short docs, paste inline; for longer ones, point them at the saved file.
3. Iterate on their feedback. Push back where their proposed change weakens the anchor; agree readily where it sharpens it.
4. Save the file at `docs/anchors/anchor-<short-title>.md` (create the directory if it doesn't exist) unless the user specifies otherwise.
5. Move to Step 4 (Verify) before declaring the anchor done.

## Step 4 — Verify

A final sanity check before the anchor is locked in. The conversation has had momentum; step out of it and read the artifact cold.

Have the user read the saved anchor back themselves, slowly. Then ask, one at a time:

- Does the problem statement still feel correct when you read it fresh?
- Has any solution language slipped in? Scan for "we should…", "add…", "build…", "use X to…".
- Does "Where it fits" actually connect to the problem you wrote, or did it drift?
- Is the bigger picture honest, or did you write something aspirational?
- Is there anything obvious you'd be embarrassed to see missing in a week?

If something is off, update the file, refresh `Last updated:`, and re-verify. Don't accept "it's fine" as a one-word answer — make the user actually read it.

When the anchor holds up under cold read, close the loop: *"Anchor verified and saved. Ready to plan how to handle it, or do you want to sit with this first?"*

## Template

Use this exact structure when generating the file.

```markdown
# anchor: <short title>

> **Last updated:** <YYYY-MM-DD>

## The problem

<One paragraph. The gap as it exists today. No solution words.>

## Where it fits

**In its immediate context:** <Where it lives, what it touches, what must stay stable around it.>

**In the bigger picture:** <The larger goal or outcome this serves. Why this, why now.>
```

## Failure modes

- **Triggered without explicit sensemaking intent.** The most common misfire. If the user didn't ask to sensemake, don't run — even if it would be useful. Sensemaking forced on someone who isn't in that mode produces theater, not understanding.
- **Solution leakage in Step 1.** Stop, restate the rule, ask again. Skipping the problem step is worse than no sensemaking at all.
- **Moving on too fast.** If the problem isn't sharply defined, don't advance to Step 2 — even if the user is impatient. A weak anchor makes the rest useless.
- **Vague "bigger picture" with no anchor.** If the user can't write Step 2's immediate context, they haven't located the work yet. Don't let them skip ahead to the bigger picture.
- **Drift into a plan.** If the spec starts describing *how* to handle the problem — steps to take, options to pick between, things to build — it's drifted into design territory. Cut those parts. This is the *make-sense-of-it* artifact, not the *how-to-handle-it* artifact.

## Handling edge cases

- **User has only a vague sense of the problem.** Resist filling in plausible-sounding language yourself. Ask more questions. The whole point is to surface what isn't decided yet. If the user remains stuck, name it — "the problem still feels thin; want to come back to this when it's clearer?" — and stop. A bad anchor is worse than no anchor.
- **User wants to revise an earlier answer mid-flow.** Always allow this. Update working notes and continue from where you were.
- **The "problem" turns out to be several problems welded together.** Say so. Suggest splitting into separate anchors. One file per anchor.
- **The user names a different save path.** Honor it. The default is a convention, not a rule.
- **The user resists the no-solution-words rule.** Hold the line gently. Explain once: "If we name the solution now, we'll stop looking for the actual problem. Give me the problem first; we can sketch the solution after the anchor is solid."

## Tone

Patient, curious, slightly stubborn. The user should feel like the questions are helping them think, not interrogating them. Push back when an answer is mushy; acknowledge clearly when an answer is sharp. The slow pace is the gift — most people don't have someone willing to sit with the problem before jumping to the solution.
