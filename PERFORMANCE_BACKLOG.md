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

- [x] Revisit full-screen string rendering after simulation-side wins.
  - Investigate whether a Bubble Tea-compatible path can avoid rebuilding a
    complete screen string every frame.
  - Stop before replacing Bubble Tea unless benchmarks show rendering dominates
    after the simulation changes.
  - Kept Bubble Tea, added sparse dirty-cell clearing, and optimized full-view
    string construction.

## Drastic Redesign

- [x] Add benchmark coverage for full frame costs.
  - Measure simulation step, sparse draw cycle, full view string construction,
    and combined model frame cost in one place.
  - Use these benchmarks to validate each larger redesign step.
  - Added `BenchmarkModelFrame` for step, draw, frame command, and `View`.

- [x] Introduce a sparse ANSI renderer.
  - Render only changed boid cells and status-line updates instead of rebuilding
    a full screen string every frame.
  - Preserve visible controls and status behavior.
  - Keep the implementation testable without requiring a real terminal.
  - Added an allocation-free sparse renderer and `BenchmarkSparseANSIFrame`.

- [x] Remove maps from sparse rendering.
  - Track dense cell state plus touched indexes to avoid hashing every boid
    and every changed cell.
  - Preserve sparse ANSI output and collision behavior.

- [x] Move interactive mode off the Bubble Tea full-view render path.
  - Use the sparse renderer for terminal output.
  - Preserve quit, pause, single-step, stats/help, reset, radius/max-speed
    runtime controls, and resize behavior where practical.
  - Added a custom raw terminal loop with sparse ANSI rendering and SIGWINCH
    resize handling.

- [x] Double-buffer simulation state.
  - Read from current arrays and write next positions/velocities, then swap.
  - Remove the full previous-frame boid snapshot from the hot path.
  - Added a reusable next-frame boid buffer and swapped buffers after each step.

- [x] Make the spatial grid fixed to simulation bounds.
  - Use screen/world dimensions to size grid arrays directly.
  - Avoid recomputing observed grid bounds every frame.
  - Track touched cells so rebuild clears only buckets used by the previous
    frame.

- [x] Reshape simulation state toward struct-of-arrays.
  - Store hot boid position and velocity state in contiguous component slices.
  - Keep behavior-compatible accessors for rendering, stats, and tests while
    avoiding per-frame struct copies.
  - Made SoA buffers authoritative, kept `Boids` as a lazy compatibility view,
    and moved stats/sparse rendering to direct SoA reads.

- [x] Gate parallel stepping for large flocks.
  - Add a fixed worker split (4 workers) only above a measured threshold (2048 boids).
  - Keep deterministic frame semantics and avoid overhead for small flocks.

- [x] Consider approximate separation math.
  - Replace per-neighbor `sqrt` only if the changed flock behavior is acceptable
    and benchmarks show a meaningful gain.
  - Kept exact `sqrt` for default-size flocks, but use deterministic inverse
    square-root approximation on large grid workloads where it improves the
    parallel branch.

- [x] Decouple simulation and render cadence.
  - Allow rendering to run at a lower or adaptive cadence than simulation when
    terminal output is the bottleneck.
  - Added `RENDER_FPS`, with simulation still stepping at `FPS` and the raw
    terminal loop rendering only when pending frames reach the render cadence.
