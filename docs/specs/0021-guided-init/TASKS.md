# Guided init tasks

## PROGRESS TABLE

| Task | State |
| --- | --- |
| Guided configuration and workflow scaffold | Implemented |
| Documentation and validation | Implemented |
| Independent review | Complete |

## TASK LIST

- [x] Add per-environment configuration and the guided command.
- [x] Generate and validate workflow scaffolding for mixed, continuous and manual environments.
- [x] Document setup, hook contracts and compatibility.
- [x] Run native tests, workflow lint and documentation checks.
- [x] Complete independent review before opening the pull request.

## TASK DETAILS

Compare the generated workflow with the v2 digest promotion contract. Require local collision tests, manual dispatch without a build path, and a shared verification job.

## DEPENDENCIES

The config type supplies validated environment modes to the workflow renderer. Complete both before testing the compiled command.

## NOTES

Provider smoke tests and Kit integration are separate.
