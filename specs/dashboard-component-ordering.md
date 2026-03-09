# Spec: Dashboard component ordering

This spec describes how dashboard components (fieldsets, entity fieldsets, actions, and other elements) are ordered in OliveTin. It documents the current behaviour so that it can be reasoned about and kept consistent.

---

## 1. Implementation

### 1.1 Two ways dashboards are built

Dashboards are built in two ways:

- **Default dashboard:** Used when there is no dashboard configuration. A single fieldset titled "Actions" is created and filled with actions that are not already on a configured dashboard.
- **Config dashboard:** Built from the dashboard configuration (e.g. under dashboards or dashboards.d). The structure is derived by walking the config tree, which produces a mix of fieldsets and a special root fieldset titled "Actions" that holds any loose items.

Ordering rules differ slightly between these two cases.

### 1.2 Top-level dashboard contents (config dashboards)

**Fieldsets without entities:** Fieldsets that are not tied to an entity type appear at the top level in **config order**. Their position in the dashboard matches the order in which they are defined in the config.

**Other top-level components:** All other top-level components (including the root "Actions" fieldset and entity fieldset groups) are **sorted** before being shown. Sort order:

1. If a component has no linked action, it is ordered by its title (alphabetically).
2. Otherwise, components are ordered first by the action's order value (lower values first).
3. If order values are equal, components are ordered by entity key: if both keys are whole numbers they are compared numerically; otherwise they are compared alphabetically.
4. If still equal, components are ordered by the action's title (alphabetically).

The root "Actions" fieldset is the single fieldset created by the build to hold loose items; it is identified by reference (not by position). When present it is added last to the list, then the sort is applied among that fieldset and entity-related components. So that fieldset can appear anywhere among those according to the rules above. When there are no loose items the root is not present, and the last component in the list is not treated as the root—so a fieldset without entities in the last position keeps config order. Regular fieldsets stay in config order and are not reordered.

### 1.3 Entity fieldsets (order of fieldsets per entity type)

When a fieldset in the config is tied to an entity type (e.g. "Server" or "Project"), one fieldset is built per entity instance. Those fieldsets are shown in **entity key order**.

**Entity key order:**

- If both keys are whole numbers: **numeric** order (e.g. 2 before 10).
- Otherwise: **alphabetical** (lexicographic) order.

So the order of entity fieldsets (e.g. one per server, one per project) is determined by this entity key order, not by config or insertion order.

### 1.4 Contents inside fieldsets

**Default dashboard:** The single "Actions" fieldset's contents are sorted. The same rules as for top-level components apply: order value first, then entity key (numeric then alphabetical), then action title.

**Config dashboards:** For all fieldsets (the root "Actions" fieldset, entity fieldsets, and regular fieldsets), the contents are **not** sorted. They keep the order from the config:

- **Root "Actions" fieldset:** Items appear in the order they are listed in the config (loose items that are not inside a fieldset).
- **Entity fieldset contents:** The order comes from the template's contents in the config.
- **Regular (non-entity) fieldset contents:** The order comes from the config, including for nested structure.

So within any config-defined fieldset, the order of actions and other child components is the **config order**.
