---
name: Precision Auction Systems
colors:
  surface: "#f7f9fb"
  surface-dim: "#d8dadc"
  surface-bright: "#f7f9fb"
  surface-container-lowest: "#ffffff"
  surface-container-low: "#f2f4f6"
  surface-container: "#eceef0"
  surface-container-high: "#e6e8ea"
  surface-container-highest: "#e0e3e5"
  on-surface: "#191c1e"
  on-surface-variant: "#44474d"
  inverse-surface: "#2d3133"
  inverse-on-surface: "#eff1f3"
  outline: "#75777e"
  outline-variant: "#c5c6cd"
  surface-tint: "#515f78"
  primary: "#000000"
  on-primary: "#ffffff"
  primary-container: "#0d1c32"
  on-primary-container: "#76849f"
  inverse-primary: "#b9c7e4"
  secondary: "#0059bb"
  on-secondary: "#ffffff"
  secondary-container: "#0070ea"
  on-secondary-container: "#fefcff"
  tertiary: "#000000"
  on-tertiary: "#ffffff"
  tertiary-container: "#002106"
  on-tertiary-container: "#0d9838"
  error: "#ba1a1a"
  on-error: "#ffffff"
  error-container: "#ffdad6"
  on-error-container: "#93000a"
  primary-fixed: "#d6e3ff"
  primary-fixed-dim: "#b9c7e4"
  on-primary-fixed: "#0d1c32"
  on-primary-fixed-variant: "#39475f"
  secondary-fixed: "#d8e2ff"
  secondary-fixed-dim: "#adc7ff"
  on-secondary-fixed: "#001a41"
  on-secondary-fixed-variant: "#004493"
  tertiary-fixed: "#83fc8e"
  tertiary-fixed-dim: "#66df75"
  on-tertiary-fixed: "#002106"
  on-tertiary-fixed-variant: "#00531a"
  background: "#f7f9fb"
  on-background: "#191c1e"
  surface-variant: "#e0e3e5"
typography:
  headline-lg:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: "700"
    lineHeight: 32px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: "600"
    lineHeight: 28px
    letterSpacing: -0.01em
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: "400"
    lineHeight: 24px
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: "400"
    lineHeight: 20px
  label-md:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: "500"
    lineHeight: 16px
    letterSpacing: 0.05em
  data-display:
    fontFamily: JetBrains Mono
    fontSize: 18px
    fontWeight: "600"
    lineHeight: 24px
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 22px
    fontWeight: "700"
    lineHeight: 28px
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  container-margin: 1rem
  stack-gap: 0.75rem
  inline-gap: 0.5rem
  section-padding: 1.5rem
---

## Brand & Style

The design system is engineered for the high-stakes environment of industrial and professional auctions. It prioritizes clarity, speed of data digestion, and an unwavering sense of institutional trust. The visual language follows a **Corporate / Modern** aesthetic with a lean toward **Minimalism** to ensure that complex bidding data remains the focal point.

The target audience consists of professional auctioneers and high-volume bidders who require real-time updates and an interface that feels like a precision tool rather than a consumer app. The UI evokes a sense of "digital headquarters"—authoritative, stable, and hyper-efficient. By stripping away ornamental elements, the design system focuses on information density and functional hierarchy.

## Colors

The palette is anchored by deep, authoritative tones contrasted against high-utility action colors.

- **Primary (#0A192F):** Used for navigation, headers, and primary text to establish institutional gravity.
- **Secondary (#007BFF):** Reserved for primary actions, active bid buttons, and interactive states.
- **Tertiary (#28A745):** Utilized exclusively for success states, active "Live" indicators, and winning bid notifications.
- **Neutral (#F8FAFC):** The primary background color to ensure a clean, breathable canvas for dense data.

Status-specific semantic colors are used for badges:

- **Pending:** #F59E0B (Amber)
- **Closed:** #64748B (Slate)

## Typography

The design system utilizes **Inter** for its exceptional legibility in UI environments and its neutral, professional tone. For data-heavy elements such as bid amounts, stock-keeping units (SKUs), and timestamps, **JetBrains Mono** is introduced to provide a distinct "data-driven" feel and ensure numerical alignment in lists.

- Use `headline-lg` for auction titles and primary dashboard headings.
- Use `data-display` for bid increments and currency values to ensure they stand out from descriptive text.
- `label-md` should be used for all status badges and micro-copy, always in uppercase when using the monospaced font.

## Layout & Spacing

This design system employs a **Fluid Grid** optimized for mobile verticality. The layout prioritizes a single-column stack to maximize the horizontal real estate for data tables and bid history.

- **Margins:** A consistent 16px (1rem) side margin is maintained across all mobile views.
- **Vertical Rhythm:** Elements are stacked using a 12px (0.75rem) gap to balance information density with touch-target accessibility.
- **Dashboards:** Key metrics are displayed in a 2-column grid at the top of the screen, transitioning into a full-width list for auction items.
- **Reflow:** On tablets, the list-detail view triggers a split-screen layout, where the list occupies 40% and the bidding interface 60%.

## Elevation & Depth

To maintain a professional and clean aesthetic, depth is communicated through **Tonal Layers** and **Low-contrast outlines** rather than heavy shadows.

- **Level 0 (Background):** #F8FAFC.
- **Level 1 (Cards/Containers):** Pure white (#FFFFFF) with a 1px solid border in #E2E8F0. No shadow.
- **Level 2 (Active/Floating):** Pure white with a very subtle, diffused shadow (0px 4px 12px rgba(10, 25, 47, 0.05)) used only for sticky bid bars or modals.
- **Separators:** 1px hairline strokes (#F1F5F9) are used to divide list items within a card.

## Shapes

The shape language is disciplined and "Soft" (4px - 8px radius). This avoids the playfulness of fully rounded corners while moving away from the harshness of sharp 0px edges.

- **Primary Buttons & Inputs:** 4px (0.25rem) radius to maintain a structural, engineered look.
- **Data Cards:** 8px (0.5rem) radius to clearly define grouped auction information.
- **Status Badges:** 2px radius or sharp edges to differentiate them from interactive buttons.

## Components

### Buttons

- **Primary:** Solid #007BFF with white text. High-contrast, 48px minimum height for mobile accessibility.
- **Secondary:** Ghost style with #0A192F border and text. Used for "View Details" or "Watchlist."

### Data Cards

- Cards must feature a "Sticky Header" area for the item name and a dedicated "Data Footer" for the current bid and time remaining.
- High-contrast background for the "Current Bid" area (use #F1F5F9) to draw the eye immediately.

### Status Badges

- **Live:** Tertiary (#28A745) background with white text. Include a pulsing dot icon for "Live" status.
- **Pending/Closed:** Subtle tinted backgrounds with dark text (e.g., light gray background for "Closed").

### Input Fields

- Structured with a persistent label in `label-md`.
- Active state indicated by a 2px #007BFF border. Use monospaced font for numeric bid entry.

### Navigation

- Bottom-fixed tab bar using #FFFFFF with #0A192F icons. The central action (typically "Search" or "My Bids") can be visually emphasized with a subtle primary color tint.
