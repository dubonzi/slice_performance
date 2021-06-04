# Improving performance with slices in Go

Go's slices are fairly easy and simple to use but when working with a lot of data, you can save both cpu and memory usage by following a few tips. I'll touch briefly on some of the details but you can read a more in depth explanation on the [official blog](https://blog.golang.org/slices-intro).

## Our example

For this example, I took the entirety of the book [Moby Dick](https://github.com/GITenberg/Moby-Dick--Or-The-Whale_2701) and split its contents to create a slice of strings (using `strings.Split(book, " ")`), which results in a array of ~190000 length. Our goal is to process each word and create a new slice of type `Word` containing our modified word.

The code used can be found on [my github](https://github.com/dubonzi/slice_performance).

Ps: Benchmarks were run using a `AMD Ryzen 5 5600X`.

```go
type Word struct {
	word  string
	index int
}
```

We then write a simple function to process our words:

```go
func ProcessWords(rawWords []string) []Word {
	words := make([]Word, 0)
	for i, w := range rawWords {
		words = append(words, Word{process(w), i})
	}

	return words
}
```
Seems fine right? Now let's make a small tweak to our function and compare their performances:

```go
func ProcessWordsFaster(rawWords []string) []Word {
	words := make([]Word, 0, len(rawWords))
	for i, w := range rawWords {
		words = append(words, Word{process(w), i})
	}

	return words
}
```
Running a benchmark for each:

```shell
BenchmarkProcessWords-12         54  21833985 ns/op   30148452 B/op  194504 allocs/op
BenchmarkProcessWordsFaster-12  100  10423307 ns/op    6086001 B/op  194471 allocs/op
```

Noticed the difference in the code? By adding a new parameter to `make`, which is the slice's initial capacity, our function now has double the performance and used around 5x less memory per operation.

Similarly to Java's ArrayList, slices are dinamically sized and backed by an array that will *grow automatically* as more elements are added. This is not a problem in most cases, but can lead to performance problems if you don't keep an eye on it. Knowing how it works is useful and can save you some resources down the road.

## How does the slice grow? 

Looking at the first function we made, when we run `make([]Word, 0)`, a new slice of type `Word` is created with an initial length and capacity of 0, which means our backing array doesn't exist yet so no memory was allocated for it.

The first time we add an element to it by using `append()`, a backing array is created with capacity 1 with our element in it. If we add another element, Go sees the array is full which means it needs to 'grow'. Arrays cannot grow, so what Go does is create a new array with an increased capacity capable of holding our new elements plus some extra space (usually doubling the capacity), then copy over everything from the previous array before adding any new elements, and pointing our slice to the new array.

This can be costly both in memory, since memory has to be allocated for the new, bigger array, and cpu, since the garbage collector will have to clean up the old array (if no other slices are pointing to it). 

Now you can imagine how growing a slice from 0 capacity to ~190000 can be costly, there will be a lot of array duplication and garbage collection happening before we get our desired result.

## Solution

The solution is simple in our case, we have an slice of n `string`s and we need a new slice of n `Word`s. By providing an initial capacity to our result array with `make([]Word, 0, len(rawWords))` we are telling Go to allocate the memory needed to hold all of our `Word`s, which in turn means the slice won't need to grow.

We are trading a little bit of overhead on the initial array creation for a big boost in performance when adding our elements.

For a little more context on the performance gains, below is the amount of times the garbage collector ran during each benchmark:

### No initial capacity

```
gc 5 @0.007s 3%: 0.004+1.1+0.006 ms clock, 0.051+0/0.93/0.74+0.081 ms cpu, 4->4->4 MB, 5 MB goal, 12 P

....

gc 373 @1.390s 3%: 0.008+0.96+0.008 ms clock, 0.10+0.007/1.5/0.74+0.096 ms cpu, 8->9->4 MB, 9 MB goal, 12 P

```

### With initial capacity

```
gc 5 @0.007s 2%: 0.004+0.49+0.004 ms clock, 0.058+0/0.88/0.59+0.055 ms cpu, 4->5->5 MB, 5 MB goal, 12 P

...

gc 174 @1.428s 1%: 0.019+1.0+0.006 ms clock, 0.23+0/1.6/0.33+0.076 ms cpu, 15->16->7 MB, 16 MB goal, 12 P

```

I've left only the first and last lines of each one since thats what we care about.

With our change, garbage collection happened 50% less.

## Conclusion

As we can see, it is possible to gain a lot of performance with a small tweak. This solution doesn't apply to all situations though, but can be adapted. What if our new slice will only contain words of 7 characters or more? We would be wasting resources by allocating memory for ~190000 words and only using a small % of that. But this is a topic for another post. 

Thank you for reading, this is my first post so feel free to provide feedback and corrections. :)
