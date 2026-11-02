---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "verify"
feature:
  id: "0017"
  slug: "linear-hotfix-requests"
  dir: "0017-linear-hotfix-requests"
---
# SPEC

## PURPOSE

Allow an authenticated Actions dispatcher to request an isolated hotfix tracked in Linear without creating a GitHub issue.

## CONTEXT

The compatibility adapter and native prepare-hotfix input require a GitHub issue. LabCore and LabCore UI now select Linear ENG. Existing workflow actor authorization and production-base/source provenance remain authoritative.

## REQUIREMENTS

Accept exactly one legacy positive Issue or canonical Linear TicketURL. Record the exact ticket URL in immutable request metadata and source PR. Use the returned ticket identifier for isolated source/base names. Preserve legacy inputs and all baseline, actor, review, fix and recovery checks. Do not add credentials, fetch Linear content, or create GitHub issues for external requests.

## ACCEPTED PLAN

Add validated external request identity to native preparation, recovery, source publication and verification. Add explicit compatibility-adapter fields for a Linear URL, baseline, fixes and reason. Document the trusted-dispatch boundary and test malformed identity, collision/recovery isolation, metadata, source verification, legacy behavior, and adapter routing. Publish through the normal reviewed upstream PR/release process before downstream adoption.

## DECISIONS

Linear is tracking metadata, not an authorization source. The existing authenticated active default-branch control run authorizes preparation; no new Linear token is needed. Reject mixing a legacy issue and external ticket. Do not fabricate a numeric GitHub identifier or use PRs as surrogate issues. Native lifecycle and provenance logic remains in Go.

## DISCOVERIES

VerifyIssue explicitly rejects pull requests. The legacy and v1.0.0 adapter parsers are identical; changing the Action pin alone cannot resolve the tracker conflict.

## VALIDATION

PASS: go test ./..., go vet ./..., make build, 26 adapter tests, and make docs-check under Python 3.13 with the pinned Markdown 3.11 dependency. Tests cover actual isolated Git source/provenance, no GitHub issue creation, invalid/mixed URLs, wrong target, cross-workspace recovery, component boundaries and authenticated-run refusal. Hosted PR checks pending.

## OUTCOME

Additive implementation and local validation complete. Legacy requests remain supported. No live hotfix request or application deployment was executed. A downstream review identified a remaining GH-only native version-allocation gate; a follow-up replaces it with shared exact marker/base validation. The new regression first failed only for Linear/scoped Linear against v1.0.1 while legacy passed, then exercises actual isolated preparation, required reviews/checks, provenance collection, tag publication to a local fixture origin, and immutable tag reuse. Upstream release and downstream pinning follow review.

## REPOSITORY MEMORY

This specification records the new external tracking boundary; public environment and adapter guides will describe supported input and failure behavior.
