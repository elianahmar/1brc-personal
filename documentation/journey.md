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
