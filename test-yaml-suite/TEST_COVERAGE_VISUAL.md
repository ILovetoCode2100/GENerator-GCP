# Virtuoso API CLI - Test Coverage Visualization

## Test Distribution Overview

```
                    VIRTUOSO API CLI TEST SUITE (335 TESTS)
                                    │
        ┌───────────────────────────┴───────────────────────────┐
        │                                                           │
        │                    Command Tests (260)                    │
        │                         77.6%                             │
        │                                                           │
        ├───────────┬───────────┬──────────┬──────────┬───────────┤
        │           │           │          │          │           │
    Interact(60)  Assert(48)  Nav(40)   Data(24)  Window(20)  Others(68)
      17.9%        14.3%      11.9%      7.2%       6.0%       20.3%
        │           │           │          │          │           │
   ╔════╧════╗ ╔═══╧═══╗ ╔═══╧═══╗ ╔══╧══╗ ╔═══╧═══╗ ╔═══╧═══╗
   ║Pos║Edg║Neg║ ║Pos║Edg║Neg║ ║Pos║Edg║Neg║ ║Pos║Edg║Neg║ ║Pos║Edg║Neg║ ║Lib║Misc║Etc║
   ║ 20║ 20║ 20║ ║ 16║ 16║ 16║ ║ 13║ 14║ 13║ ║ 8 ║ 8 ║ 8 ║ ║ 7 ║ 6 ║ 7 ║ ║ 24║ 28║ 16║
   ╚═══╩═══╩═══╝ ╚═══╩═══╩═══╝ ╚═══╩═══╩═══╝ ╚═══╩═══╝ ╚═══╩═══╩═══╝ ╚═══╩═══╩═══╝

                              Additional Tests (75)
                                    22.4%
        ┌───────────────┬──────────────┬──────────────┬───────────────┐
        │               │              │              │               │
    Workflows(20)    YAML(15)    Errors(30)    Performance(10)
        6.0%           4.5%         9.0%           3.0%
```

## Command Group Details

### Step-Interact (60 tests) 🔥

```
click ███████████████ (15)
double-click █████ (5)
right-click █████ (5)
hover ████████ (8)
write ██████████ (10)
key █████ (5)
mouse ███████ (7)
select █████ (5)
```

### Step-Assert (48 tests) ✅

```
exists ██████ (6)
not-exists ████ (4)
equals ██████ (6)
not-equals ████ (4)
checked ███ (3)
selected ███ (3)
variable ████ (4)
gt/gte/lt/lte ████████████ (12)
matches ██████ (6)
```

### Step-Navigate (40 tests) 🧭

```
to ██████████ (10)
scroll-top ███ (3)
scroll-bottom ███ (3)
scroll-element █████ (5)
scroll-position ████ (4)
scroll-by █████ (5)
scroll-up █████ (5)
scroll-down █████ (5)
```

## Test Type Distribution

```
          Test Types Across All Commands

    Positive Tests  █████████████████████████ 33.3%

    Edge Cases     █████████████████████████ 33.3%

    Negative Tests █████████████████████████ 33.3%
```

## Coverage Heatmap

```
        Low Coverage  │  Medium Coverage  │  High Coverage
              🔵      │       🟡         │      🟢
    ─────────────────┴──────────────────┴────────────────

    🟢 step-interact    60 tests  ████████████████████
    🟢 step-assert      48 tests  ████████████████
    🟢 step-navigate    40 tests  █████████████
    🟡 step-data        24 tests  ████████
    🟡 step-window      20 tests  ███████
    🟡 step-dialog      20 tests  ███████
    🟡 library          24 tests  ████████
    🟡 workflows        20 tests  ███████
    🟡 yaml-features    15 tests  █████
    🟢 error-scenarios  30 tests  ██████████
    🔵 step-wait         8 tests  ███
    🔵 step-file         8 tests  ███
    🔵 step-misc         8 tests  ███
    🔵 performance      10 tests  ███
```

## Execution Time Estimates

```
    Test Group          │ Tests │ Avg Time │ Total Time
    ────────────────────┼───────┼──────────┼───────────
    Command Tests       │  260  │   3.5s   │  15.2 min
    Workflows           │   20  │  30.0s   │  10.0 min
    YAML Features       │   15  │   2.0s   │   0.5 min
    Error Scenarios     │   30  │   1.5s   │   0.8 min
    Performance         │   10  │  45.0s   │   7.5 min
    ────────────────────┼───────┼──────────┼───────────
    TOTAL               │  335  │    -     │  34.0 min
```

## Priority Matrix

```
    Priority Level     Tests    Execution Strategy
    ┌───────────────────────────────────────────────┐
    │ 🔴 Critical (Smoke)     50 tests   Always run │
    │ 🟠 High                 120 tests   Daily      │
    │ 🟡 Medium               100 tests   Weekly     │
    │ 🟢 Low                   65 tests   Monthly    │
    └───────────────────────────────────────────────┘
```

## Success Metrics

```
    Metric                Target    Current    Status
    ───────────────────────────────────────────────────
    Command Coverage      100%      100%       ✅
    Test Count            300+      335        ✅
    Execution Time        <30m      ~34m       ⚠️
    Flaky Rate            <1%       TBD        🔄
    Pass Rate             >95%      TBD        🔄
```
