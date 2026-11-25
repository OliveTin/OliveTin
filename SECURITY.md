# Security Policy

## Supported Versions

The following branches are currently being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| `main` (3k release branch)  | :white_check_mark: - advisories will be published when patched in this branch |
| `release/2k` (2k release branch) | :x: - end of life, no security fixes or bug fixes (see below) |

To understand more about 2k vs 3k, see the following docs; https://docs.olivetin.app/upgrade/2k3k.html

### Policy change (25th November 2026): end of support for OliveTin 2k

Until 25th November 2026, this policy said OliveTin 2k would receive bug fixes and security updates until at least 31st December 2028. That has changed: **as of 25th November 2026, OliveTin 2k is end of life, and no longer receives any updates - neither security fixes nor bug fixes.** We're sorry. The original commitment was made in good faith, and we know some of you planned around it.

There are four reasons for the change:

* **There isn't enough maintainer time to patch 2k.** In practice, 2k has received almost no patches since 3k was released. All the available time has gone into 3k, and it is more honest to say so than to keep a promise on paper that isn't being kept.
* **Very few people are asking for 2k.** Interest in staying on 2k has been very low compared to the number of users who have simply upgraded to 3k.
* **Many fixes can't be made safely in 2k's architecture.** Recent fixes in 3k rely on typed entities, the Connect RPC API, and a configuration system that can be reloaded cleanly. 2k has none of these, so a lot of fixes would mean rewriting large parts of 2k, which would effectively rebuild 3k inside 2k.
* **Reports are arriving much faster than 2k could be patched.** AI-assisted security research now finds issues faster than ever. That is good for OliveTin users overall, but it means 2k fixes would ship late, incomplete, or not at all, and a branch that still claims security support without getting it is worse than an honest end-of-support notice.

What this means for you:

* `release/2k` no longer receives security fixes or bug fixes, and new advisories are only assessed against 3k.
* Existing 2k releases and container images (`latest-2k`) remain available, and nothing will stop working.
* If you are on 2k, please plan your move to 3k. See the upgrade guide: https://docs.olivetin.app/upgrade/2k3k.html. If something blocks your upgrade, please open an issue. Fixing upgrade blockers is now the best way we can keep 2k users safe.

## OliveTin *is* a remote code execution (RCE) "tool"

The very purpose of OliveTin is to allow users to execute commands remotely on a machine.

This means that, by design, OliveTin has much higher potential to be used for remote code execution (RCE), and any security vulnerabilities that do occur have the potential to be much more severe than in other types of software.

We hope that you understand that while the project goes to great aims to be safe, and mitigate, that security vulnerabilities are inevitable, as they are with all software of all sizes - like Kubernetes, the Kernel, etc - and OliveTin has substantially less resources than those projects.

With that being said, OliveTin tries to follow examples of best practice, so judge the project not on if/when it has security issues, but how security issues are responded to as the measure of quality.

This is why we take security very seriously, and why we encourage responsible disclosure practices when reporting vulnerabilities.

## Reporting a Vulnerability

Please use responsible disclosure practices when reporting a vulnerability. **You will receive full credit for your discovery**, and we will work with you to ensure that the issue is resolved as quickly as **possible**. Please note that only James Read has access to security issues at the moment, so please be patient and understanding if you do not receive an immediate response.

* **Option A (preferred)**: GitHub Security Advisories, which allows you to report a vulnerability privately and securely. Use this direct link to report privately: `https://github.com/OliveTin/OliveTin/security/advisories/new`. This allows you to provide details without making them public.

* **Option B**: Please email `contact@jread.com` for responsible disclosure.

The following notes might be helpful when reporting a vulnerability:

* OliveTin does not offer a bug bounty program.
* GitHub usernames are how we you will be credited for discoveries reported via GitHub, if using emails we'll ask for your preferred name/handle to credit you with.
* CVEs will be requested via GitHub Security Advisories when appropriate, but we do not guarantee that all vulnerabilities will receive CVEs, as this is determined on a case-by-case basis. See "When CVEs are requested" below.
* Please state the exact version or commit you reproduced the issue on.

## When CVEs are requested

A CVE is only requested when the vulnerability can be reproduced on at least one of:

* the latest published OliveTin 3k release, or
* the current `next` branch.

Reports that only reproduce on older releases, and are already fixed in the latest 3k release or on `next`, are closed without a CVE. If a published advisory already covers the fix, the report is closed as a duplicate of that advisory, and the reporter is still credited.

Reports that only affect OliveTin 2k are not eligible for a CVE (see "End of support for OliveTin 2k" above).

Reports that cannot be reproduced on the latest 3k release or `next` may be closed until they can.

## Feature flags (alpha / experimental)

OliveTin uses global `features.*` flags in `config.yaml` to ship unfinished or experimental functionality.

* **All feature flags default to off.** Enabling a flag is an explicit operator choice.
* Functionality behind a feature flag is **alpha / experimental** until the flag is removed or the feature is graduated to a stable, default-on product surface.
* Private security reports **are accepted** for vulnerabilities that affect feature-flagged (alpha) functionality when that flag is enabled. Use Option A or B above; do not file a public issue that discloses exploit details.
* Reports that affect **stable, non-flagged** code paths remain in scope under this policy, even if a feature flag exists elsewhere in the project.

Operators who enable experimental features should treat them as preview software and avoid relying on them in high-assurance production deployments.

## Disclosure of how vulnerabilities were found

It is incredibly useful to not just patch security vulnerabilities, but also to understand how they were found. If you are able to share this information, it can help us and the community to better understand potential attack vectors and improve the overall security of the project.

## Duplicate reports

If you are reporting via GitHub Security Advisories, search existing [repository advisories](https://github.com/OliveTin/OliveTin/security/advisories) for the same component and attack path before filing. Maintainers may close duplicate submissions and continue work on a single canonical advisory; duplicate reporters are still credited.

Maintainers: see [.github/SECURITY_ADVISORY_DUPLICATES.md](.github/SECURITY_ADVISORY_DUPLICATES.md) for known duplicate pairs, triage steps, and OAuth2 issues that are easy to confuse with each other.

## Denial of service (DoS) reports

Any web service can be overwhelmed by enough traffic, so not every way of making OliveTin slow or busy is a vulnerability. Rate limiting request floods is the job of a reverse proxy, load balancer or firewall in front of OliveTin, not of OliveTin itself.

A DoS report **is** considered a vulnerability when at least one of these is true:

* **Crash or hang**: one request, or a small number of requests, crashes or hangs the OliveTin process. For example a panic, a nil pointer dereference, a fatal Go runtime error such as concurrent map writes, or a deadlock.
* **Amplification**: a small or cheap request makes OliveTin do disproportionately large work. For example a single request that allocates a large amount of memory, or a short input that causes a large amount of CPU work.
* **Unbounded growth**: requests cause memory or other state to grow without limit, and it is not reclaimed after the requests stop. For example a map, slice or cache that has no cap, expiry or eviction.

A DoS report is **not** considered a vulnerability when:

* the load is proportional to the number and size of requests sent, and the work done for each request is bounded;
* OliveTin fails gracefully while busy, for example by returning HTTP 429 or `ResourceExhausted`, rather than crashing;
* OliveTin recovers by itself once the requests stop;
* the only possible fix would be rate limiting inside OliveTin.

When a DoS report is accepted, the preferred fix is to put a bound on the resource (a concurrency limit, a size cap, an expiry or eviction, or a limit on the cost of evaluating input), not to add rate limiting. Severity is assessed on how practical the attack is: an issue that needs a very large volume of requests before it has an effect will usually be scored lower than the reporter's estimate.

## Known behaviour that is not considered a vulnerability

The following have been reported several times and reviewed. Please only report them if you have found something beyond what is described here.

* **`PasswordHash` is unauthenticated and can make local login busy.** The `PasswordHash` API is a documented utility for generating Argon2id hashes, and is needed before any user exists. Unbounded resource use was fixed in CVE-2026-28342 by limiting all Argon2id work, including `LocalUserLogin`, to 10 concurrent operations. Flooding `PasswordHash` can make logins return `ResourceExhausted` while the flood continues; this fail-fast behaviour is intentional, keeps memory bounded, and has no lasting effect. Rate limit request floods at your reverse proxy.
* **`DEBUG` and `TRACE` logs contain sensitive information.** The `DEBUG` and `TRACE` log levels, `-startuptrace`, and `logDebugOptions` exist to show OliveTin's internal workings while troubleshooting. They are expected to log sensitive information, such as configuration values including secrets, API keys, password hashes, OAuth2 client secrets, webhook secrets, and request headers. This is by design and is not considered a vulnerability. Treat logs produced with these options as being as sensitive as `config.yaml` itself, and only enable them while troubleshooting. Reports about sensitive information in logs are only in scope at the default `INFO` level, or at `WARN` and `ERROR`.

## Process

Once a vulnerability is reported, the process is;

* Check [.github/SECURITY_ADVISORY_DUPLICATES.md](.github/SECURITY_ADVISORY_DUPLICATES.md) and open advisories for duplicates before accepting.
* Confirm the report reproduces on the latest published 3k release or the `next` branch before accepting it or requesting a CVE.
* Accept or reject the report, and communicate with the reporter about next steps.
* If accepted, patch using a temporary branch, and code review will be requested from the original reporter if they are interested.
* The severity of the vulnerability will be assessed using CVSS, and the patch will be prioritised accordingly.
* Once the patch is ready, it will be queued for a release onto the `next` branch (3k).
* The reporter will be credited in the advisory and the release notes, but not the commit message.
* The commit message will contain a reference to the CVSS score (eg: MED) and the advisory ID.
