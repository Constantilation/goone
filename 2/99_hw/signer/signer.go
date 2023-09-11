package main

import (
	"sort"
	"strconv"
	"sync"
)

var mu = &sync.Mutex{}

var jobCaller = func(job job, wg *sync.WaitGroup, in, out chan interface{}) {
	defer func() {
		wg.Done()
		close(out)
	}()
	job(in, out)
}

var ExecutePipeline = func(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	for _, job := range jobs {
		wg.Add(1)
		go jobCaller(job, wg, in, out)
		in = out
		out = make(chan interface{}, 100)
	}

	wg.Wait()
}

type LogicFunction = func(value string) string

var IterateOverChan = func(in, out chan interface{}, function LogicFunction) {
	wg := &sync.WaitGroup{}
	for value := range in {
		switch convertedInterface := value.(type) {
		case string:
			wg.Add(1)
			go func() {
				defer wg.Done()
				out <- function(convertedInterface)
			}()
			break
		case int:
			wg.Add(1)
			go func() {
				defer wg.Done()
				out <- function(strconv.Itoa(convertedInterface))
			}()
			break
		default:
			continue
		}
	}

	wg.Wait()
}

var SingleHash = func(in, out chan interface{}) {
	var singleHash = func(value string) string {
		crcValueChannel := make(chan string)
		crcMd5Channel := make(chan string)

		go func() {
			crcValueChannel <- DataSignerCrc32(value)
		}()

		go func() {
			mu.Lock()
			md5Result := DataSignerMd5(value)
			mu.Unlock()
			crcMd5Channel <- DataSignerCrc32(md5Result)
		}()

		crcMD5Value := <-crcMd5Channel
		crcValue := <-crcValueChannel

		return crcValue + "~" + crcMD5Value
	}

	IterateOverChan(in, out, singleHash)
}

var MultiHash = func(in, out chan interface{}) {
	th := 6

	var multiHash = func(value string) string {
		var hashesSlice = make([]string, 6)

		wg := &sync.WaitGroup{}

		for i := 0; i < th; i++ {
			wg.Add(1)
			i := i
			go func() {
				defer wg.Done()
				hashesSlice[i] = DataSignerCrc32(strconv.Itoa(i) + value)
			}()
		}

		wg.Wait()
		var sum string
		for _, val := range hashesSlice {
			sum += val
		}

		return sum
	}

	IterateOverChan(in, out, multiHash)
}

var CombineResults = func(in, out chan interface{}) {
	var resultSlice = make([]string, 0)

	for value := range in {
		switch convertedInterface := value.(type) {
		case string:
			resultSlice = append(resultSlice, convertedInterface)
		case int:
			resultSlice = append(resultSlice, strconv.Itoa(convertedInterface))
		default:
			continue
		}
	}

	sort.Strings(resultSlice)

	var resultHash string
	for index, hash := range resultSlice {
		if index == 0 {
			resultHash += hash
		} else {
			resultHash += "_" + hash
		}
	}

	out <- resultHash
}
