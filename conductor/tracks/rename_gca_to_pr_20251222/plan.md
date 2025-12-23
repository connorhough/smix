# Plan: Rename `gca review` to `pr review`

## Phase 1: Preparation and Renaming [checkpoint: c518478]
- [x] Task: Create `cmd/pr.go` by copying/refactoring `cmd/gca.go` 514c548
- [x] Task: Update `cmd/pr.go` to use `pr` as the command name and `review` as the subcommand 514c548
- [x] Task: Register `prCmd` in `cmd/root.go` and remove/deprecate `gcaCmd` 514c548
- [x] Task: Update `internal/gca/` references in `cmd/pr.go` if necessary 514c548
- [x] Task: Conductor - User Manual Verification 'Phase 1: Preparation and Renaming' (Protocol in workflow.md) c518478

## Phase 2: Documentation and Cleanup [checkpoint: 053aba1]
- [x] Task: Update `README.md` examples from `smix gca review` to `smix pr review` 4d035d9
- [x] Task: Update any other documentation or comments referring to `gca review` 38068bc
- [x] Task: Remove old `cmd/gca.go` file 12629cc
- [x] Task: Conductor - User Manual Verification 'Phase 2: Documentation and Cleanup' (Protocol in workflow.md)

## Phase 3: Internal Package Refactoring
- [ ] Task: Rename `internal/gca/` directory to `internal/pr/`
- [ ] Task: Update package declarations and imports to use `pr`
- [ ] Task: Update internal string references and folder naming logic
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Internal Package Refactoring' (Protocol in workflow.md)
