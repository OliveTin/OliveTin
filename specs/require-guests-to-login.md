# Spec: Requiring guests to log in

This spec describes what visitors who are not logged in ("guests") may do when the installation requires guests to log in.

---

## 1. Scope

The rules apply only when the installation is configured to require guests to log in. When it is not, guests get the default permissions and any access-control entries that match them, as usual.

A guest is a visitor whose request carries no valid credentials of any kind. A logged-in user is never a guest, even if their username or usergroup is "guest".

---

## 2. API calls guests may make

Guests may only make the API calls needed to log in and to check that the server is running:

- the initial client bootstrap, which tells the client that login is required;
- the readiness check;
- local user login;
- logout;
- generating a password hash.

Every other API call made by a guest is refused with "permission denied". This includes streaming calls, the calls that start, restart and kill actions, and the diagnostic calls that the installation enables with its insecure options.

The refusal happens before the request is looked at. A guest gets the same answer for an action, execution or entity that exists as for one that does not, so guests cannot find out which ones exist.

The list of calls guests may make is closed. Any API call added in future requires login unless it is added to this list on purpose.

---

## 3. Permissions

Guests have no permissions at all:

- all default permissions are off;
- no access-control entry applies to a guest, even one that matches the "guest" username or usergroup.

Access-control entries that match the "guest" usergroup still apply to logged-in users in that usergroup, for example users from an identity provider that gives them that group.

This is enforced separately from the API call rule in section 2, so that a guest cannot start, view, or kill an action even through a path that does not go through the API.

---

## 4. Logging

When an action is requested, the log records the username and the authentication provider of the user who requested it, at the default log level, so that actions requested by guests can be told apart from actions requested by logged-in users.
