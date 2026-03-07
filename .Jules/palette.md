## 2026-02-26 - Notification Bell Accessibility
**Learning:** Icon-only buttons with dynamic badges (like "unread count") are often inaccessible. The visual badge is separate from the icon, leading to confusing screen reader output (e.g., "Notifications, 3").
**Action:** Use a dynamic `aria-label` on the parent container (e.g., `aria-label="Notifications (3 unread)"`) and hide the visual elements (icon and badge) with `aria-hidden="true"`. This provides a single, coherent announcement.

## 2026-02-26 - Keyboard Visibility for Hover Interactions
**Learning:** Elements that appear on `hover` (like quick actions on a card) are invisible to keyboard users. Simply adding `opacity-100` on focus isn't enough if the container has `pointer-events-none` (often used to prevent clicks on invisible elements).
**Action:** Use `group-hover:opacity-100 focus-within:opacity-100` to show the element. Crucially, pair `pointer-events-none` on the base state with `group-hover:pointer-events-auto focus-within:pointer-events-auto` to ensure the controls become interactive when they are revealed, either by mouse or keyboard focus.

## 2026-02-26 - Disconnected Form Labels
**Learning:** React elements with isolated `<label>` and `<input>` tags, where the label does not explicitly wrap the input, require `htmlFor` and `id` bindings to function. Without this, clicking the label fails to focus the input (a huge usability regression), and screen readers cannot properly associate the label with the input.
**Action:** Always verify that `<label>` tags use `htmlFor` matching the associated `<input>` or `<textarea>`'s `id`, especially when inputs are rendered conditionally or separated by sibling DOM elements (like a wrapping div). For dynamic forms mapped from a list, always use unique IDs using the item's key (e.g., `id={\`edit-note-\${link.id}\`}`).
