/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"github.com/jrallison/go-workers"
	"sync"
)

func configureWorkers() {
	workers.Configure(map[string]string{
		"server":   "localhost:6379",
		"database": "0",
		"pool":     "30",
		"process":  "1",
	})
}

func main() {
	configureWorkers()
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(2)

	var j int
	for j = 0; j < 100; j++ {
		go func() {
			for _ = range jobs {
				workers.Enqueue("benchmark_worker", "Add", []string{"hello", "hello2"})
				wg.Done()
			}
		}()
	}

	for i := 0; i < 2; i++ {
		jobs <- 1
	}

	wg.Wait()

}
