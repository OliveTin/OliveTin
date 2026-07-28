# Spec: Dashboard navigation categories

This spec describes how root dashboards can be grouped into categories in the sidebar navigation.

---

## 1. Configuration

Root dashboard entries in the configuration may include an optional category label.

- The category applies only to root dashboards (top-level items in the dashboards list). Nested dashboard contents ignore category.
- If category is omitted or empty, the dashboard is uncategorized.
- Dashboards that share the same category label are grouped together under that label in the sidebar.

## 2. Visibility

Only dashboards the current user is allowed to view appear in navigation.

- Access-denied dashboards are omitted from the list and do not create empty category sections.
- If every dashboard in a category is hidden, that category does not appear.

## 3. Sidebar ordering

When building the sidebar:

1. Uncategorized dashboards appear first, as a flat list above any category sections, in configuration order among visible uncategorized dashboards.
2. Category sections follow, in the order each category first appears among visible categorized dashboards.
3. Within a category, dashboards keep the order they appear in the configuration among visible dashboards in that category.
4. After all dashboard links, a **System** category lists Entities, Logs, and Diagnostics (each only when the user is allowed to see that item). If none of those links are visible, the System category is omitted.

The default Actions dashboard, when present, is uncategorized unless configuration gives it a category (it is not a configured root entry, so it stays uncategorized).

## 4. Navigation style

Category sections apply when section navigation uses the sidebar. Top-bar navigation does not show category section headers; dashboard links still appear in the same relative order without collapsible category groups.
