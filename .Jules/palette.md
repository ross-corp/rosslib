## 2026-02-26 - Notification Bell Accessibility
**Learning:** Icon-only buttons with dynamic badges (like "unread count") are often inaccessible. The visual badge is separate from the icon, leading to confusing screen reader output (e.g., "Notifications, 3").
**Action:** Use a dynamic `aria-label` on the parent container (e.g., `aria-label="Notifications (3 unread)"`) and hide the visual elements (icon and badge) with `aria-hidden="true"`. This provides a single, coherent announcement.

## 2026-02-26 - Keyboard Visibility for Hover Interactions
**Learning:** Elements that appear on `hover` (like quick actions on a card) are invisible to keyboard users. Simply adding `opacity-100` on focus isn't enough if the container has `pointer-events-none` (often used to prevent clicks on invisible elements).
**Action:** Use `group-hover:opacity-100 focus-within:opacity-100` to show the element. Crucially, pair `pointer-events-none` on the base state with `group-hover:pointer-events-auto focus-within:pointer-events-auto` to ensure the controls become interactive when they are revealed, either by mouse or keyboard focus.

## 2026-03-02 - Accessible Icon Buttons in Lists
**Learning:** When using icon-only buttons with dynamic content (like vote toggles) in lists, the `aria-label` must reflect the current state (e.g., 'Remove upvote' vs 'Upvote') rather than just the general action.
**Action:** Always check if the `title` or visual state of an icon button changes dynamically and mirror that logic in the `aria-label`.
