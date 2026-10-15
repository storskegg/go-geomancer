# Plan 04 — Suggesting the Significator from the Question

Follows plan 03 (PDF report), which adds the question prompt. This milestone uses the
question to suggest the significator: the significator prompt opens on the best match, so a
correct suggestion needs only Enter.

Everything runs offline from data built into the binary: no network, API keys or tokens.

## Flow

1. `cast` asks for the question (from plan 03; free text, skippable).
2. The question is scored against the twelve houses.
3. The significator prompt opens on the top suggestion and describes the top two:

   > Suggested: House 7 · Partnership (from "divorce") · or House 2 · Wealth (from "money")

   The user accepts with Enter or chooses another. With no question, or no confident
   suggestion, it opens on House 1 as now.
4. `--significator` still skips the prompt; with it, the suggestion isn't used.

## Approach: lexicon plus word vectors

Two scorers, combined. Both are pure Go.

### Lexicon

- Per house, weighted terms: House 7 has *marry, husband, wife, partner, divorce, contract,
  lawsuit, thief, stolen, …*; House 10 has *job, promotion, boss, career, reputation, …*.
- Kept in the content file beside the significations (`content/data/en.json`), so it can be
  translated with them.
- Words are stemmed with the Snowball English stemmer (`github.com/kljensen/snowball`, BSD),
  so *marrying* and *married* both match *marri*. Common words (*will, I, my, get, …*) are
  dropped.
- A match records the term, so suggestions can say why.

### Word vectors

- Pre-trained GloVe vectors (Stanford, public domain under the PDDL), trimmed to the most
  common English words (about 40k) and quantized to int8: a few MB, embedded.
- Each house has a prototype vector: the weighted mean of its lexicon terms and its
  signification's words. A question's vector is the mean of its content words' vectors,
  and is compared with each prototype by cosine similarity.
- This catches words the lexicon doesn't list: *fiancée*, *betrothal* and *wedding* land near
  House 7.
- A build-only tool (`internal/tools/vectors`, not part of the binaries) trims and quantizes
  the downloaded GloVe file. The output is committed, so builds don't need the download.

### Combining

- Score = lexicon score + λ × vector similarity, with λ tuned on the test set.
- Suggestions are the top two houses, with a confidence; the second is shown only if it's
  close to the first. Below a confidence threshold, no suggestion is made.

## Code

- New core package `question`: `Suggest(text string) []Suggestion`, where a `Suggestion` has a
  house, a score and the matched terms. No dependency on the CLI, so the future API and web
  app can use it.
- `chart.Reading` gains `Question string` (from plan 03) and the suggestions aren't stored:
  they're derived from the question.
- The significator prompt takes an initial house and a description of the suggestions.

## Evaluation

- `question/testdata/questions.tsv`: about 200 labelled questions, covering each house with
  varied phrasing, plus deliberately ambiguous ones labelled with two acceptable houses.
- A test reports accuracy and pins thresholds (to set from the first results, aiming for
  top-1 ≥ 85% and top-2 ≥ 95%), and lists misclassifications when it fails.

## Phase 2: derived houses

Questions about another person's affairs use the house counted from that person's house:
"Will my husband get the job?" is the 10th from the 7th, House 4.

- Detect who the question is about (*husband/wife/partner* → 7, *sister/brother* → 3,
  *son/daughter/child* → 5, *father* → 4, *mother* → 10, *friend* → 11, *boss/employer* → 10, …)
  and the subject (*job* → 10).
- Derived house = ((person − 1) + (subject − 1)) mod 12 + 1.
- Offer it as a suggestion with its derivation: "House 4 · the 10th from the 7th (husband's
  career)". Labelled test questions cover it.

## Out of scope

- Languages other than English.
- Transformer models: the only pure-Go option (`nlpodyssey/cybertron`, on `spago`) was last
  released in November 2023, and the others need cgo.

## Outcome

Built as planned, phase 2 included.

- `question.Suggest` scores the lexicon (671 terms across the twelve houses, in
  `content/data/en.json` with the significations) and the GloVe vectors (50,000 words ×
  100 dimensions, int8, 5.4 MB in `question/data/vectors.bin`, built by
  `internal/tools/vectors`).
- Lessons from tuning, now rules: terms sharing a stem are one term; phrases take their words
  first; a word that is itself a term isn't claimed by another term sharing its stem
  ("faithful" isn't "faith"); only terms of weight 2+ and words outside the 2,000 most common
  are matched by meaning; derivations need an exact subject match that isn't a quality
  (`qualities`: "faithful", "loyal", …), isn't owned by the querent ("my work"), and isn't in a
  first-person question. A subject may be in the person's own house ("Will my boss be fired?"
  is the 10th from the 10th).
- No suggestion means the prompt opens on House 1, so evaluation scores it as House 1.

### Accuracy

| Set | Questions | Top 1 | Top 2 | Notes |
|---|---|---|---|---|
| `questions.tsv` | 186 | 97.8% | 98.9% | tuned on |
| `holdout.tsv` | 48 | 93.8% | 93.8% | first measured at 66.7% / 72.9%; the lexicon was then expanded by category in response, so it's no longer unseen |
| `holdout2.tsv` | 40 | 75.0% | 82.5% | written afterwards and never tuned on: the realistic figure for new questions |

The test pins each set a little below these. Misses on new questions are mostly vocabulary the
lexicon lacks (audition, treaty, blood pressure); adding terms by category is the way to
improve, measured against a fresh held-out set each time.

### Interface

- The significator prompt starts on the top suggestion and shows the top two as a hint, e.g.
  "Suggested: ④ Home, the 10th from the 7th (“husband”, “job”); or ⑩ Honour, from “job”".
  It suggests again only if the question changes, so a house chosen by hand survives going
  back. Accessible prompts show the hint in the title and accept it with Enter.
- `geomancer suggest QUESTION` (and `--json`) shows the suggestions and their reasons.
