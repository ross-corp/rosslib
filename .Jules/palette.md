## 2026-02-26 - Notification Bell Accessibility
**Learning:** Icon-only buttons with dynamic badges (like "unread count") are often inaccessible. The visual badge is separate from the icon, leading to confusing screen reader output (e.g., "Notifications, 3").
**Action:** Use a dynamic `aria-label` on the parent container (e.g., `aria-label="Notifications (3 unread)"`) and hide the visual elements (icon and badge) with `aria-hidden="true"`. This provides a single, coherent announcement.
