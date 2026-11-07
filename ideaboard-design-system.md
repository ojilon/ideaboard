# Ideaboard — Design System Specification

**Status:** Concept specification  
**Scope:** Desktop-first note-taking canvas with a responsive Android companion  
**Visual direction:** Calm, minimal, paper-like, friendly, and focused. Use soft surfaces and restrained pastel accents without sacrificing text contrast or control clarity.

This specification turns the Ideaboard visual concepts into a coherent set of design decisions. Values are proposed design tokens, not measured values from a production app. The canvas is the primary workspace; controls should remain quiet and notes should carry the visual emphasis.

## 1. Design principles

1. **Canvas first.** Keep persistent interface chrome light and compact. Let notes, regions, and connections provide the information hierarchy.
2. **Quiet contrast.** Use warm near-whites and charcoal ink rather than stark white and pure black. Pastels are for note surfaces and small accents, not for essential text.
3. **Structure without weight.** Regions use thin dashed outlines; connectors use fine strokes; cards use subtle borders and restrained shadows.
4. **Small, consistent geometry.** Prefer 4–6 px control corners and a compact 4 px spacing base. Larger radii are reserved for floating panels and mobile toolbar containers.
5. **Readable at a glance.** Note text should remain comfortably legible at typical zoom. Labels and controls must not rely on color alone.
6. **Interaction is explicit.** Selected, focused, hovered, and disabled states should be distinguishable with outline, fill, or opacity changes in addition to color.

## 2. Color system

### 2.1 Light theme tokens

| Token | Hex | Intended use |
|---|---:|---|
| `canvas` | `#F7F8F6` | Default board background |
| `surface` | `#FFFFFF` | Settings panels, dialogs, toolbar surfaces |
| `surface-muted` | `#F2F4F1` | Secondary controls, quiet card areas |
| `surface-hover` | `#EEF2EE` | Hovered controls and list rows |
| `ink-primary` | `#27313A` | Primary text and icons |
| `ink-secondary` | `#66717A` | Supporting labels and metadata |
| `ink-tertiary` | `#89939A` | Placeholder text and low-priority metadata |
| `line` | `#DCE2DE` | Borders, dividers, control outlines |
| `line-strong` | `#BCC7C0` | Focus-adjacent structure, selected outlines |
| `region-outline` | `#B8C6BD` | Dashed board-region boundaries |
| `accent-sage` | `#789A89` | Primary accent, active tools, confirm actions |
| `accent-sage-soft` | `#DCE9E0` | Soft selected/active background |
| `accent-blue` | `#7895C5` | Secondary connector or selection accent |
| `accent-lavender` | `#B5A6D8` | Board and note accent |
| `accent-peach` | `#E8B9A9` | Board and note accent |
| `accent-yellow` | `#F2D977` | Board and note accent; quick-capture selection |
| `focus-ring` | `#507D68` | Keyboard focus indicator |
| `danger` | `#B85C55` | Destructive actions and errors only |
| `success` | `#668D71` | Success status and completed actions |

**Surface hierarchy:** `canvas` is the largest uninterrupted surface. Use `surface` for floating UI. `surface-muted` should support controls without becoming another competing panel. Avoid pure black and high-chroma fills.

### 2.2 Note palette

Sticky-note backgrounds are intentionally soft. Use primary ink for body copy and avoid white text on pastel fills.

| Note token | Fill | Border / accent | Recommended use |
|---|---:|---:|---|
| `note-yellow` | `#FFF4C7` | `#E8D991` | New ideas, quick capture |
| `note-sage` | `#E2EEE3` | `#B9D0BD` | Plans, actions, grounded thoughts |
| `note-peach` | `#F8E5DE` | `#E7C5B8` | Prompts, people, open questions |
| `note-lavender` | `#EEE8F8` | `#D0C3E8` | References, reflection, creative notes |
| `note-blue` | `#E5EDF8` | `#C4D2E8` | Research and supporting context |
| `note-neutral` | `#F1F2EF` | `#D9DEDA` | General-purpose notes |

Color should not encode meaning without a label, icon, or other cue. Users can choose a note color freely; the suggested meanings above are optional defaults only.

### 2.3 Dark theme tokens

The dark theme uses a warm charcoal canvas and muted pastel note fills. Avoid a pure-black canvas and neon highlights.

| Token | Hex | Intended use |
|---|---:|---|
| `canvas` | `#202127` | Main dark board background |
| `surface` | `#292B33` | Floating panels and toolbars |
| `surface-muted` | `#30333C` | Secondary control background |
| `surface-hover` | `#393D47` | Hovered controls and rows |
| `ink-primary` | `#ECEEF0` | Primary text and icons |
| `ink-secondary` | `#B4BAC2` | Supporting labels and metadata |
| `ink-tertiary` | `#8C949E` | Placeholder and low-priority text |
| `line` | `#41454E` | Borders and dividers |
| `line-strong` | `#5B626D` | Stronger control outlines |
| `region-outline` | `#555B64` | Dashed region boundaries |
| `accent-sage` | `#9EBEAA` | Active tool and primary accent |
| `accent-sage-soft` | `#35463D` | Active/selected soft background |
| `accent-blue` | `#9BB1D7` | Connector or selection accent |
| `accent-lavender` | `#C5B8E5` | Board and note accent |
| `accent-peach` | `#E8BDB0` | Board and note accent |
| `accent-yellow` | `#F0D985` | Board and note accent |
| `focus-ring` | `#B1D0BC` | Keyboard focus indicator |
| `danger` | `#E18B82` | Destructive actions and errors |
| `success` | `#A4C5A9` | Success status |

Dark-theme note fills can remain pastel, as in the concept images, but should be slightly desaturated and paired with dark readable note text. Do not map light-theme body text directly onto pastel cards without checking contrast.

### 2.4 Transparency and background images

- Default panel opacity: **94–98%** over plain canvas; use full opacity for small text-heavy panels when needed for contrast.
- Board background-image overlay: apply a neutral tint or dimming layer so region labels and notes remain clear. The image must never compete with note content.
- Prefer a blur of approximately **2–4 px** only when the user explicitly chooses a photo background and readability requires it. Avoid blurring the notes or toolbar.
- Transparency sliders should show a visible track, thumb, and numeric value; preserve sufficient contrast at every slider position.

## 3. Typography

### 3.1 Typeface direction

Use a friendly modern sans-serif for all interface labels and most note text. Recommended stacks:

- **UI:** `Inter`, ` system-ui`, `-apple-system`, `BlinkMacSystemFont`, `"Segoe UI"`, sans-serif.
- **Handwritten-feeling note option:** a restrained, legible handwritten face such as `Kalam` or `Caveat`, with sans-serif fallback. Use only for note content, not controls, settings, installer text, or long passages.
- If a custom font is unavailable, keep the same size and line-height with a system sans-serif fallback; do not substitute a highly decorative typeface.

### 3.2 Type scale

| Style | Size / line height | Weight | Typical use |
|---|---|---:|---|
| Display | 32 / 40 px | 600 | Board-list page title, if needed |
| Page title | 24 / 32 px | 600 | Board or dialog title |
| Section title | 18 / 26 px | 600 | Settings headings, panel sections |
| Body large | 16 / 24 px | 400–500 | Sticky-note text, primary mobile labels |
| Body | 14 / 20 px | 400 | Standard control labels and cards |
| Body medium | 13 / 18 px | 500 | Toolbar labels, selected states |
| Caption | 12 / 16 px | 400 | Edited dates, helper text |
| Micro label | 11 / 14 px | 500 | Compact metadata; use sparingly |

- Use sentence case for labels and settings.
- Use a maximum of two weights within a compact control group.
- Prefer normal or medium weights; reserve semibold for titles and key actions.
- Default note text is **16 px / 24 px** on desktop and **16 px / 22–24 px** on mobile. Allow user zoom to scale note text with the canvas.
- Line wrapping should preserve short thought fragments. Avoid overly narrow default note widths.
- Text underline is a deliberate annotation treatment: a fine 1–2 px line below one selected word or phrase. Do not use underline as a background highlight or as a general link style on notes.

## 4. Spacing and layout

### 4.1 Spacing scale

Use a 4 px base grid. Choose values from this scale for component padding and gaps:

| Token | Value | Typical use |
|---|---:|---|
| `space-0` | 0 px | Reset |
| `space-1` | 4 px | Icon-label gap, tiny internal spacing |
| `space-2` | 8 px | Compact control padding, small gaps |
| `space-3` | 12 px | Tight card/control padding |
| `space-4` | 16 px | Standard component padding |
| `space-5` | 20 px | Toolbar and panel spacing |
| `space-6` | 24 px | Card padding, section gap |
| `space-8` | 32 px | Main layout gaps |
| `space-10` | 40 px | Board-region separation |
| `space-12` | 48 px | Large section separation |
| `space-16` | 64 px | Canvas safe edge / broad whitespace |

### 4.2 Desktop workspace

- Standard desktop composition is 16:9, but the actual application should adapt to the available viewport.
- Maintain a **24–40 px** minimum canvas inset from the window edge, with **40–64 px** preferred on wide displays.
- Keep app identity and board title in a compact top header. Recommended header height: **56–64 px**.
- Floating toolbars should be **44–52 px** high, with **8 px** internal padding and **8–12 px** gaps between icon buttons.
- Settings and inspector panels should be compact, with **20–24 px** internal padding and **20–28 px** between sections.
- Dotted region outlines should have a minimum interior padding of **24 px** from their label or note content.
- Leave enough negative space between note clusters to make connectors and regions immediately distinguishable.
- For board lists, use a fluid grid with **16–24 px** gutters. Target three columns on wide desktop and two columns at narrower desktop widths.

### 4.3 Mobile workspace

- Use **16–20 px** side margins and preserve a thumb-friendly bottom toolbar.
- The top app bar should be around **56 px** high; a bottom toolbar should be about **64–72 px**, excluding device safe-area inset.
- Maintain a minimum **44 × 44 px** touch target for primary tools and menu actions.
- Keep canvas content at its natural scale and support pan/zoom; do not shrink an entire desktop board to fit the phone screen.
- Region outlines may stack vertically or appear partially in view; each should remain recognizable and independently navigable.
- Respect Android safe areas and gesture insets in implementation. Keep primary creation tools accessible from the bottom edge.

## 5. Shape, borders, and elevation

### 5.1 Corner radius

| Token | Value | Usage |
|---|---:|---|
| `radius-xs` | 3 px | Tiny labels or badges |
| `radius-sm` | 5 px | Buttons, text fields, toolbar buttons |
| `radius-md` | 6 px | Sticky notes, board cards, small panels |
| `radius-lg` | 10 px | Settings panels and dialogs |
| `radius-pill` | 999 px | Segmented controls or compact status chips, only when appropriate |

Keep the visual language mostly square-soft rather than pill-heavy. Note cards may vary subtly in tilt, but core controls should remain aligned and stable.

### 5.2 Border and line weights

- Standard UI border: **1 px** solid `line`.
- Selected control: **1 px** `accent-sage` border plus soft `accent-sage-soft` fill.
- Focus indicator: **2 px** `focus-ring`, with at least **2 px** offset where space allows.
- Region outline: **1–1.5 px** dashed, with a calm, consistent dash pattern.
- Connector: **1.5–2 px**; dotted connectors use a clearly visible dot rhythm. Arrowheads should be small and precise.
- Hand-drawn annotations may be slightly irregular, but should remain thinner and visually subordinate to note content.

### 5.3 Shadows

Use shadows sparingly to distinguish floating UI from the canvas, not to create depth everywhere.

- Note: `0 2px 6px rgba(35, 45, 40, 0.08)`.
- Toolbar: `0 4px 16px rgba(35, 45, 40, 0.08)`.
- Panel/dialog: `0 8px 28px rgba(35, 45, 40, 0.12)`.
- In dark mode, prefer a 1 px border and subtle surface separation over a heavy black shadow.

## 6. Core component specifications

### 6.1 Sticky note

- Default desktop size: approximately **180–240 px wide × 112–168 px high**, content-driven within that range.
- Padding: **16–20 px**; body text 16/24 px; minimum line count one.
- Use one note fill from the note palette. Optional one-word underline should use a slightly stronger related accent.
- Default state has no visible resize handles. On selection, show a fine border, small resize handles, and an unobtrusive contextual toolbar.
- Preserve a usable text-editing area at mobile zoom; do not scale text below 14 px in an active note.

### 6.2 Board region

- Use a thin dashed rounded outline with a short label such as a weekday or topic.
- Label may sit just inside the upper-left boundary with **12–16 px** inset.
- Region labels use Body / 14 px medium or Body large / 16 px medium when board zoom is high.
- Region boundaries must remain visually distinct from connectors and background image detail.

### 6.3 Floating toolbar

- Desktop height: **44–52 px**; surface white/light or dark equivalent; 1 px border; subtle shadow.
- Tool order: selection/pen, note, region, image, connector (adapt per context).
- Icon-only controls need accessible names and hover/focus tooltips. Do not rely on icon shape alone where ambiguity is likely.
- Active tool: soft accent fill plus stronger icon color; inactive tools use primary or secondary ink.
- Quick-capture toolbar should appear near the active note but avoid obscuring typed content.

### 6.4 Settings panel

- Suggested desktop width: **320–380 px**.
- Group controls by theme, font, background image, and transparency, using clear labels and consistent vertical spacing.
- Prefer segmented controls for a two-way Light/Dark choice, a select/dropdown for font, a button plus preview for background image, and a labeled slider with a numeric value for transparency.
- Use a visible close affordance and preserve keyboard navigation order.

### 6.5 Board card

- Use a quiet preview area, readable title, and low-priority edited timestamp.
- Suggested desktop card width in a three-column grid: **280–440 px**, depending on viewport. Keep preview ratio consistent (around 16:9 or 3:2).
- Make the whole card a clear navigation target while preserving a distinct overflow-menu target.
- Avoid dense metadata and unnecessary category badges.

### 6.6 Installer/dialog pattern

- Center a focused dialog with a clear app name and concise task title.
- Path field, Browse action, progress bar, status, and primary action should each be visually distinct.
- Show progress in a restrained accent color and pair it with a text status; never communicate progress only through color.
- Installer concepts are separate from the in-app canvas and may use a stronger dialog hierarchy.

## 7. Interaction and state rules

| State | Treatment |
|---|---|
| Default | Calm surface, standard border, primary ink |
| Hover | Subtle `surface-hover` fill; no layout movement |
| Focus-visible | Clear 2 px focus ring; retain readable text |
| Selected | Accent border plus soft accent fill or selection handles |
| Pressed | Slightly darker soft fill; avoid scaling or bouncing |
| Disabled | Reduced contrast and opacity, but keep label discernible |
| Error | `danger` text/icon plus a short explanation; never color alone |
| Success | `success` plus clear status copy where confirmation matters |

- Animations should be restrained: **120–180 ms** for hover and control state changes; **180–240 ms** for panels or popovers.
- Use ease-out for opening and ease-in for closing. Avoid springy or elastic motion in the core workspace.
- Respect reduced-motion preferences. Do not animate connectors or note positions without direct user action.
- Canvas dragging should not accidentally start text selection. Note editing and object movement need distinct interaction modes.

## 8. Accessibility and legibility

- Target WCAG AA contrast for all essential interface text and controls. Check muted text against its actual surface, including transparent panels over photos.
- Do not use low-contrast pastel text on pastel note backgrounds; use `ink-primary` for note copy.
- Maintain visible keyboard focus and logical tab order across header, toolbar, canvas objects, and panels.
- Provide accessible names for icon-only controls; expose tool and selected states to assistive technology.
- Support keyboard equivalents for create note, select tool, pan canvas, zoom, and open settings where applicable.
- Use a minimum **44 × 44 px** touch target on mobile, even when the visible icon is smaller.
- Ensure selected notes and regions are indicated through more than color alone.

## 9. Responsive behavior

- **Wide desktop:** Keep the header compact, float the tool palette near the top center or active selection, and use broad canvas whitespace.
- **Compact desktop/tablet:** Reduce outer padding, let panels overlay the canvas, and avoid squeezing the board into multiple narrow columns.
- **Phone portrait:** Use a compact header and bottom toolbar; allow vertical canvas scrolling/panning and gesture-based zoom. Region content may flow vertically.
- **Board list:** Three columns on wide desktop, two on medium widths, and a single column on phone. Preserve card title and edited date before decorative preview detail.
- **Photo background:** Reduce image contrast or add a translucent neutral veil automatically when readable contrast falls below target.

## 10. Content and writing style

- Keep note examples short, conversational, and idea-oriented: “Start small”, “Ask Maya”, “Sketch a plan”.
- Use sentence case for board names, settings, and actions.
- Prefer verbs for actions: “Choose image”, “New board”, “Install”.
- Keep helper text direct and brief. Avoid overly formal or productivity-heavy language.
- Use underlines as a small expressive accent beneath an individual word, not as a broad highlight bar.

## 11. Starter implementation checklist

- [ ] Define semantic light and dark color tokens and test contrast on both plain and photo backgrounds.
- [ ] Create shared type styles and verify note text at desktop and phone zoom levels.
- [ ] Apply the 4 px spacing scale and the 5–6 px control radius consistently.
- [ ] Establish reusable note, region, toolbar, panel, board-card, and dialog patterns.
- [ ] Verify keyboard focus and touch target sizes for every toolbar action.
- [ ] Test mobile board navigation without shrinking the desktop canvas to an unreadable overview.
- [ ] Confirm that selection and status are never conveyed by color alone.
