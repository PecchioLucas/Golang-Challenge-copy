# Golang-Challenge
Challenge test

We ask that you complete the following challenge to evaluate your development skills.

## The Challenge
Finish the implementation of the provided Transparent Cache package.

## Show your work

1.  Create a **Private** repository and share it with the recruiter ( please dont make a pull request, clone the private repository and create a new private one on your profile)
2.  Commit each step of your process so we can follow your thought process.
3.  Give your interviewer access to the private repo

## What to build
Take a look at the current TransparentCache implementation.

You'll see some "TODO" items in the project for features that are still missing.

The solution can be implemented either in Golang or Java ( but you must be able to read code in Golang to realize the exercise ) 

Also, you'll see that some of the provided tests are failing because of that.

The following is expected for solving the challenge:
* Design and implement the missing features in the cache
* Make the failing tests pass, trying to make none (or minimal) changes to them
* Add more tests if needed to show that your implementation really works
 
## Deliverables we expect:
* Your code in a private Github repo
* README file with the decisions taken and important notes

## Time Spent
We suggest not to spend more than 2 hours total, which can be done over the course of 2 days.  Please make commits as often as possible so we can see the time you spent and please do not make one commit.  We will evaluate the code and time spent.
 
What we want to see is how well you handle yourself given the time you spend on the problem, how you think, and how you prioritize when time is insufficient to solve everything.

Please email your solution as soon as you have completed the challenge or the time is up.

## SOLUTION

# Thinking process :
* Started by reading the problem and getting to know the problem and the given partial-solution: 15-20 minutes.

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/4860600: First I started looking for a way to have the cache price creation time somewhere and the best way I found was to use a struct that contained the price along with every extra fields that I needed, in this case only the creation time. Now that I had this struct I needed to add the check if the element was expired or not. For that I had 2 options, one was to add a method for TransparentCache the other was to add a method for CachedPrice new struct. I didn't know what kind of problems I was going to face when starting to write down the parallelization of the routines so I wrote it as a method of TransparentCache just to be sure that if I needed to change the behaviour I had all the information that the TransparentCache has. Now I think it's better as a method of CachedPrice: isExpired(maxAge). 

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/76a359b: Having solved the cache expiration the next step was to start solving the parallelization problem. I knew that the cache was going to need some mutex to avoid conflicts during the concurrent/parallel access to the resource. I chose RWMutex that provides 2 different locking methods, RLock() for reading and just Lock() for writing. RWMutex is faster than the Mutex when reading as the RLock() doesn't lock other reading routines, only writing ones.

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/39124d2: I added the parallelization using go routines, a waitGroup to wait for all the routines to be processed and a few channels (prices and errors) but quickly realized that with that approach it wasn't going to be easy to tell what itemCode threw which error. If I kept on that approach I would have had to add some context to the errors to be more specific when returning them. The advantage of having a struct is that if I needed something else from the routine processing result I could have just added a new field into it without changing too much the rest of the code.

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/5900da5: Performed the channel changes as it was explained in the last commit explanation.

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/c6188b2: In this commit I added the missing processing of the routines results to the expected slice of prices. And whenever there was an error retrieving one of the prices for an itemCode, I created a simple error describing the amount of errors ocurred. At this point, all tests were OK including the one that was failing at first because of the parallelization feature.

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/317ad89: I added a new unit test for the scenario of the getPrices method returning error for some of the itemCodes.

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/97ce066: I found it to be a good improvement to have a maximum amount of routines running at the same time. The implementation could have had a number of parallel routines to access the cache and a different number of parallel routines trying to access the slow PriceService but for that I needed to change the original GetPrice() too much and I had little time left. So I decided to add a semaphore as a buffered channel of struct{}. By having this buffered channel and adding one empty struct just before it started processing and releasing it as soon as it was over processing, I had the semaphore I needed to have only the amount of routines permitted as the routines were going to wait for the buffered channel to release an element to avoid the channel overflow. I could have done this with golang default semaphore library but this is intended for cases in which you need to assignt different weights to different routines. Also I used empty structs and not ints because empty structs size is 0 bytes. 

* Commit https://github.com/PecchioLucas/Golang-Challenge/commit/b7aa33c: Having very little time left I decided to take advantage of the private struct I've created. I slightly improved the error handling so that an error is returned as a list of errors causes along with the itemCode that caused it.

# Time spent: 

* 15-20 mins understanding the problem and the given partial-solution.
* aproximately 1 hour 45 minutes in the problem solving.
* 1 hour 30 minutes writing this README.