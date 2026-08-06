# Spec: Entity access control

This spec describes how OliveTin restricts which entity types a user may see, and how that interacts with entity-related actions.

---

## 1. Scope

Access control applies at the **entity type** level (each configured entity definition), not per instance.

- Instances remain data loaded from entity files.
- Only the ability to **view** an entity type is consulted for these rules.
- Separate action permissions (view, execute, logs, kill) continue to govern actions themselves.

---

## 2. Configuration

Each entity definition may list zero or more named access-control entries.

- If the list is omitted or empty, the entity type is **unrestricted**: any user who may use the dashboard UI can see that type (same rule as root dashboards with no access-control list).
- If one or more named entries are listed, access is an allow list: a matching entry that grants view, otherwise the installation’s default view permission.
- Marking an access-control entry as applying to every action does not apply to entity types. An entry must be listed on the entity definition to restrict it.

---

## 3. Listing and details

When listing entity types and instances:

- Types the user cannot view are omitted from the list. The response does not advertise that those types exist.
- A request for a single entity whose type the user cannot view is treated as not found (the type is not confirmed to exist).

The Entities navigation page remains available; it simply shows fewer (or no) types when some are restricted.

---

## 4. Search hints

Client search hints for entities include only instances of types the user may view.

Entity-bound action hints (actions generated per entity instance) appear only when the user may view both the action and the entity type. Action view alone is not enough if the entity type is restricted.

Search hints are omitted from the initial client bootstrap when guests must log in, when header search is disabled for the installation, and the header search control is not shown until login is no longer required and header search is enabled.

Hints are capped at 100 actions and 50 entity instances **per entity type** per bootstrap response. The client applies the same caps when indexing.

Dashboards are not included in search hints. Clients build the dashboard search index from the bootstrap root dashboard entries (already filtered by dashboard access control).

Entity types with no access-control list remain unrestricted for search and listing.

---

## 5. Entity-related actions

There are two shapes of related actions on an entity details page:

1. Actions bound to the entity type (one binding per instance).
2. Unbound actions that prefill arguments from the entity.

Rules:

- Opening the entity details page requires view on that entity type.
- Seeing or running a related action still requires the action’s own permissions.
- Entity type access does not replace action permissions.

---

## 6. Dashboards and argument forms

When a dashboard expands an entity fieldset, instances of types the user cannot view are not rendered.

When an action argument draws choices from an entity type, it must define **exactly one** choice template and name the entity type. OliveTin expands that template per instance. Only instances of types the user may view are included. Users who can view the action but not the entity type must not learn instance names from the argument form.

Arguments that name an entity type with zero or multiple choice templates are invalid configuration: startup/reload rejects them, Diagnostics reports an error, the argument form shows no choices, and start/validate requests are rejected.

Starting or validating an action rejects entity-backed argument values when:

- The user may not view that entity type (permission denied), or
- The value is not one of the entity-expanded choice values for that argument (invalid argument).

Guessing an instance name must not bypass entity type access control.

---

## 7. Compatibility

Existing entity definitions without access-control lists stay unrestricted. Setting the default view permission to false alone does not hide unrestricted entity types; operators must list access-control entries on entity definitions to lock them down.
