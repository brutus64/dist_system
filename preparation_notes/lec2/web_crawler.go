package main

import (
	"fmt"
	"sync"
)

type cache struct {
	seen map[string]bool
	mu sync.Mutex
}
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, c *cache, wait *sync.WaitGroup) {
	defer wait.Done() //when goroutine done reading url, (this includes 1st one from beginning call)
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	
	for _, u := range urls {
		//lock before reading cache
		c.mu.Lock()
		if _, exists := c.seen[u]; !exists {
			//not seen so mark as seen and crawl it
			//goroutine needs a channel or sync.WaitGroup to have Crawl wait for it to finish before returning to main, otherwise its just calling goroutine but returning right away since its not blocking
			c.seen[u] = true
			wait.Add(1) //need to wait for another crawl
			c.mu.Unlock()
			go Crawl(u, depth-1, fetcher, c, wait)
			//fmt.Println(c.seen)
		} else {
			//seen skip it
			c.mu.Unlock()
		}

	}
	return
}

func main() {
	var wait sync.WaitGroup
	c := &cache{ seen:make(map[string]bool) } //need to initialize a cache, mutex doesnt require initalization
	wait.Add(1)
	c.seen["https://golang.org/"] = true
	Crawl("https://golang.org/", 4, fetcher, c, &wait)
	wait.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}