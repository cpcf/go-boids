# Performance Backlog

Goal: improve runtime performance enough to be visible in benchmarks and worth
manual testing in the terminal.

## Current Priority

- [x] Replace the map-backed spatial grid with a dense linked-cell grid.
  - Reduce neighbor-search overhead in `simulation.Step`.
  - Preserve exact neighbor filtering and existing behavior.
  - Verified with `rtk go test ./...` and focused update benchmarks.

- [x] Remove square-root/map work from render-direction selection.
  - Replace direction normalization plus `map[Point]rune` lookup with a small
    direct direction calculation.
  - Preserve the visible eight-direction triangle behavior.
  - Verified with `rtk go test ./...` and focused benchmarks.

- [ ] Revisit full-screen string rendering after simulation-side wins.
  - Investigate whether a Bubble Tea-compatible path can avoid rebuilding a
    complete screen string every frame.
  - Stop before replacing Bubble Tea unless benchmarks show rendering dominates
    after the simulation changes.
