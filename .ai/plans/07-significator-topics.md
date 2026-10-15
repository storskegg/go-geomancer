# Plan 07 — What the Question Asks About Within the House

A house signifies several matters: House 6 is sickness, but also work, servants and small
animals. When the significator is a house the question points to, the reading names the part of
the house the question is about, not just the house's title: "Should I get a new pet?" gives
"Significator: House 6, Small animals" rather than "House 6, Illness". When it can't tell (the
house was chosen by hand, or there's no question), it uses the title as before.

## Decisions

| Topic | Decision |
|---|---|
| Topics | Each house's lexicon terms are divided among named topics following its text: House 6 is `sickness`, `habits`, `work`, `servants`, `small animals`. Every term is in exactly one topic. Names are lowercase and don't contain "and", which joins two topics in a label |
| Content file | `significations[].topics: [{name, terms}]` replaces `terms`; `Signification.Terms` is still filled in, as the union, when the file is parsed |
| When it's known | Only when the chosen house is one of the question's suggestions. Its topics are those of the terms its words matched (directly or by meaning), ranked by the score they contributed. A derived house takes the topic of the subject term in the subject's own house |
| Label | The top two topics, capitalized and joined with "and": "Small animals", "Sickness and small animals". Derived: "Husband's career (the 10th from the 7th)" |
| Where | Replaces the title in the PDF header and reading headings, the terminal's significator line and reading headings, the prompt's suggestion hint and `geomancer suggest`. The terminal line keeps the house's full text after the dash |
| Stored | `Reading.Matter` (`matter` in JSON: topics, and the derivation if any), set by `cast` after the significator; `SetSignificator` clears it |

## Code

- `content.Topic`, `Signification.Topics`, `Signification.Topic(term)`.
- `question.Suggestion.Topics`; `Suggestion.Matter()`; `Suggester.Matter(q, house)`.
- `chart.Matter` with `Subject()` (for headings) and `Label()` (with the derivation, for the
  significator line); `chart.Derivation` moved from `question`, which aliases it, so `chart`
  doesn't import `question` and its word vectors; `chart.Ordinal`.

## Not done

- Topics don't change suggestion scores; accuracy is unchanged.
- The house texts weren't rewritten to match the topic names exactly (House 2's "gain and loss"
  is the topic `trade`; House 1's "body, health" is `health`).
