# Context

Languages:
- Go

We will develop two application based on the same common core logic:
- a pretty cli application
- an api server with future intentions of...
    - serving a reactjs web application
    - being the backbone of a more native application (electron, flutter, whatever)

# What We're Doing

Developing a Geomancy application capable of calculating multiple, selectable geomantic techniques (e.g. traditional or golden dawn). This requires customizable mappings per "System" below.

Create a plan to build this project in Go with the information found in this prompt.

# Ontologies

## System

I know of at least 2 major systems used for geomancy:
1. "Traditional"
2. "Golden Dawn"

For the sake of our ontologies, a system is an interfaced set of mappings between bodily and elemental correspondences of the Segments of a Tetragram, and the House-placement of each tetgragram in the astrological square. Additional differences may need to be included later, but this is a good starting point.

## Tetragrams

The core of the system involves 16 Tetragrams. These Tetragrams are made up of 4 Segments. Tetragrams are grouped together in following categories:
- Mothers: There are 4 Mothers. These are provided by the user.
- Daughters: There are 4 Daughters. These are calculated from the Mothers.
- Nieces: There are 4 Nieces. These are calculated from both the Mothers and the Daughters. Whether they're called Nieces or Nephews is a matter of preference. I prefer Nieces, but a user may have their own preference.

These 12 Tetragrams are what get mapped to the astrological square's houses based on the chosen System's mapping. We can call them the 12 principal Tetragrams.

### Additional Meta

The following meta is important:
- The number of dots, total, in a tetragram
- The number of even dots in a tetragram (the qty of even segments * 2)
- The number of odd dots in a tetragram (the qty of odd segments)
- Corresponding element
- Corresponding astrological Sign
- Corresponding astrological Planet
- Description (Long)
- Description (Stort/Qualitative)
- Name
- Qualitative correspondence with each of the Houses

### Math

Furthermore, Tetragrams can be added together. The sum of two Tetragrams is calculated by comparing the Segments of the two Tetragrams being added, and follows the same basic rules of arithmetic when it comes to adding odd/even numbers:
- An Odd and an Even always equals an Odd
- Two Odds or Two Evens always equals an Even

For example...

M1 = {Head: 1, Neck: 2, Body: 1, Feet: 2}
M2 = {Head: 1, Neck: 1, Body: 2, Feet: 2}

M1 + M2 = {Head: 2, Neck: 1, Body: 1, Feet: 2}

## Segments

Segments are binary representations of "odd" or "even." In a Tetragram, each Segment corresponds with a part of the body, top to bottom:
1. The Head
2. The Neck
3. The Body/Torso
4. The Feet

Additionally, the bodily attributions also have elemental correspondences, which vary by the chosen System's mapping.

Odd is represented by a single dot.
Even is represented by two dots.

Odd segments are considered "active."
Even segments are consistered "passive."

## Spread

There are two primary spreads, and both will always be calculated:
1. The first is called "The Shield," and is alwys the starting point
2. The second is called an astral or astrological square.

### The Shield

The Shield is how we get started. It neatly displays the following information:
- The Mothers
- The Daughters
- The Nieces
- The Right and Left Whitnesses
- The Judge

### The Astrological Square

The Astrological Square shows more detail. It arranges the 12 principal Tetragrams into their respective Houses, based on the chosen System's mapping, in a square format. Each House has a Zodiacal Sign that gets calculated from the Tetragram in the 1st House. Each House also has an optional elemental correspondence calculated backwards from the active segments of the Judge Tetragram.

## Groups

As mentioned, the 12 principle tetragrams are organized into 3 groups:
- The Mothers (4) -- M1, M2, M3, M4
- The Daughters (4) -- D1, D2, D3, D4
- The Nieces (4) -- N1, N2, N3, N4

Also as previously mentioned, The Mothers are provided by the user. The reset, however, are calculated.

### The Daughters

The Daughters are calculated from the Mothers as follows:

D1:
Head: M1.Head
Neck: M2.Head
Body: M3.Head
Feet: M4.Head
D2:
Head: M1.Neck
Neck: M2.Neck
Body: M3.Neck
Feet: M4.Neck
D3:
Head: M1.Body
Neck: M2.Body
Body: M3.Body
Feet: M4.Body
D4:
Head: M1.Feet
Neck: M2.Feet
Body: M3.Feet
Feet: M4.Feet

The daughters are the only tetragrams to be calculated in this way. Every subsequent tetragram will be calculated by addition (alreadty described how they're summed).

### Nieces

The Nieces are calculated from the Mothers and Daughters as follows:

N1: M1 + M2
N2: M3 + M4
N3: D1 + D2
N4: D3 + D4

### Special Tetragrams

The Right and Left Whitnesses are a special group of special tetragrams. They are calculated as follows:

WR: N1 + N2
WL: N3 + N4

The Judge (J) is calculated as WR + WL
The Judgement (JM) is calculated as J + M1

