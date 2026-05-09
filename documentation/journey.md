## First day 5/2/26
- I pretty quickly implemented the logic for parsing the file
- For just reading in the file line by line, it takes 27.68 seconds
- Okay, now, I'm trying to convert each line into a measurement object. (5-2-oom.txt
- Worked for an hour on this. Wrapping up for today. Next, I will need to implement the solution on a small subset of the data. Ensure correctness


## Second day 5/3/26
- To read 100 million lines it takes 91.58 seconds
- added a validation method to ensure correctness at the end




## Third day 5/6/26
- I refactored the heck out of the code
- Code is ready to run. Time to get a baseline
- When running the application I'm seeing that memory from activity monitor reached as high as 38GB
- Got a baseline of 473 seconds (7.88 minutes)
- I noticed I'm off by .1. Specifically I'm predicting .1 under for all of the misses across min, max, and avg
- I just implemented the producer consumer pattern where two go routines are running concurrently, with one pushing to a channel and the other receiving from the channel and updating a map. This brought the runtime down to 426 seconds. -47 seconds. Good improvement. 413 to go!
- I added some helpful metrics to track misses across min, max, and avg. Additionally, I also added metrics for tracking the number of cities processed. We only process 413 cities
  - This means that a majority of the time, actually comes from reading in the file. Which is obvious in retrospect, but helpful nonetheless to know.
  - Now, I can say with certainty that a majority of the performance gains will come from reading the file concurrently


## Fourth day 5/6/26
- Finally got the memory/cpu profiling working. Ran the full profiling and am finally getting some good profiling data back. 
- From heap profile. Almost all the time is going into the ReadFile() function which confirms my suspicion
- For the cpu profiling I'm seeing the following

```txt
Showing top 10 nodes out of 64
      flat  flat%   sum%        cum   cum%
    59.85s 18.15% 18.15%     59.85s 18.15%  runtime.memclrNoHeapPointers
    57.79s 17.53% 35.68%     57.79s 17.53%  runtime.usleep
    39.31s 11.92% 47.60%     39.31s 11.92%  runtime.madvise
    35.09s 10.64% 58.25%     36.55s 11.09%  runtime.scanobject
    31.53s  9.56% 67.81%     31.53s  9.56%  countbytebody
    30.82s  9.35% 77.16%     30.82s  9.35%  runtime.memmove
    28.70s  8.70% 85.86%     28.70s  8.70%  syscall.syscall
    20.81s  6.31% 92.17%     20.81s  6.31%  runtime.pthread_cond_signal
     6.84s  2.07% 94.25%      7.61s  2.31%  runtime.mallocgcTiny
     2.73s  0.83% 95.07%      2.73s  0.83%  runtime.(*mspan).init
(pprof) %
```


### Fifth Day 5/8/26
- Finally got a concurrent implementation for file reading working
- However, the runtime is significantly worse. I'm getting 631 seconds
- From the heap dump, I'm seeing that reconcileChunks() eating my memory. This was the function I created because, reading by chunk, leads to cases where I a line could be cut off

### 6th Day 5/9/26
- Okay, did some more profiling and saw that reading the file concurrently took 11 seconds. That's blazing fast. SO I think the main concern now, is that I need to intelligently correct lines at the boundary point.
- My overall performance for the concurrent implementation became worse. But I think I will need to optimize how I bring the chunks together. I have some ideas but will have to see. Maybe at the offset, I could splitN instead
