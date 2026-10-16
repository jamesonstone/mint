# PLAN

1. Preserve existing conventional commit parsing and GitHub reference precedence.
2. Accept only explicit canonical Linear URLs from full footer lines. Render identifier links alongside commit links and any GitHub issue link.
3. Keep generation offline; do not add provider authentication or ticket-content enrichment. Document public visibility and unavailable provider behavior.
4. Test valid, duplicate, malformed and mixed references, run repository checks, and deliver an unmerged PR coordinated with Kit GH-264.

No release or production mutations are in scope.
