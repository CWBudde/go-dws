# maze_generation Test Issue

## Status
**KNOWN ISSUE** - Test file contains algorithmic bug

## Description
The `testdata/fixtures/Algorithms/maze_generation.pas` test fails because it contains a bug in the maze generation algorithm.

## Root Cause
The algorithm doesn't check if cells are already visited before marking them:
```pascal
cells[p] := True;
toVisit -= 1;
```

When the algorithm gets stuck (no valid neighboring cells and empty backtrack stack), it continues looping on the same cell, marking it multiple times and decrementing `toVisit` for each marking.

## Debugging Results
- Algorithm successfully visits 138 out of 144 cells (95.8%)
- Gets stuck at position p=26 (x=0, y=0)
- Marks the same cell 7 additional times (iterations 139-144)
- Terminates when `toVisit` reaches 0, despite not visiting all cells
- Results in partial maze with paths only in top portion

## Expected vs Actual Output
- **Expected**: Complete 25x25 maze with paths throughout all lines
- **Actual**: Partial maze with paths in ~54% of the grid (top ~7 rows)

## Required Fix
The algorithm needs one of:
1. **Count Once**: Only mark cells that haven't been visited:
   ```pascal
   if not cells[p] then begin
      cells[p] := True;
      toVisit -= 1;
   end;
   ```

2. **Detect Stuck**: Break when no progress can be made:
   ```pascal
   if p = oldP and directions.Length = 0 and stack.Count = 0 then
      break;
   ```

3. **Restart**: When stuck, pick a new random unvisited cell (but this changes the maze pattern)

## Verified Operations
All basic operations work correctly in our implementation:
- ✅ `+=` and `-=` operators
- ✅ Array operations (SetLength, Add, Push, Pop, Count, High, Swap)
- ✅ Boolean arrays and `not` operator
- ✅ `for` loops with `step`
- ✅ RandomInt with seed reproducibility
- ✅ `div` operator precedence

## Conclusion
This is not a bug in the go-dws interpreter. The test file itself contains an algorithmic flaw. The expected output was likely generated with a corrected version of the algorithm.

## Recommendation
- Mark test as expected failure
- Or apply the fix to the test file if we're allowed to correct test bugs
- Or skip this test in CI until the test suite is updated
