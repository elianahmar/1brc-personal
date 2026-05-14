### TODO
- Benchmark all of the parts of my program. 
  - Benchmark the file reading
  - Decompose the different parts down in the benchmark so I can compare performance
- Convert floating point to ints and do conversion once at the end
- Have final validation being written as a json object. So I can jq over it
- Fix the under by .1 issue. Thinking it's an issue with truncating
- Aim for under 15 seconds
- Add unit tests for the simple parser. I don't think it's working quite right

### Constraints
- Can only use std library packages
- No ai under any circumstances. It's for my own learning
- Document each and every step along the way

### DONE
- Read the entire file into memory -> DONE
- Calculate the min max and average of 1 billion measurements -> DONE
- I need to disable copilot by default when using neovim. Copilot LSP is using gigs of memory 
- Add helpful metrics to track misses and cities processed
- Fix pprof not giving any data for the CPU profiling
- See if I can profile a script using pprof
- Parallel file reading
  - This will require me to add synchronization around the map. For now, just use a mutex
- At the last line, add total number of misses, min misses, avg misses, max misses
- multiple concurrent consumer pattern. Not really helping. I think I need to see the dumps
- Write files to directories based on day. Put under documentation folder
- add .gitignore
- Objectify everything so that swapping out different implementations become easier
- Add support for command line args
- Utilize unsafe for zero allocations. Halved my runtime
- Benchmark bufio.Reader versus scanner (using ReadSlice('\n')) -> DONE
