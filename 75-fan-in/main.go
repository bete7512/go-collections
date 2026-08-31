package main

import "sync"

func main() {}
func Merge(chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, ch := range chans {
		if ch == nil{
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func MergeCollect(chans ...<-chan int) []int {
	out := Merge(chans...)
	results := []int{}
	for v := range out {
		results = append(results, v)
	}
	return results
}

type Labeled struct {
	Source string
	Value  int
}

func MergeWithLabels(sources map[string]<-chan int) []Labeled {
	results := []Labeled{}
	for k, ch := range sources {
		for v := range ch {
			results = append(results, Labeled{Source: k, Value: v})
		}
	}
	return results
}
