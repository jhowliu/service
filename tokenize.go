package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

func buildRequest(method, host, path string) *http.Request {
	uri := host + path
	req, _ := http.NewRequest(method, uri, nil)

	return req
}

func queryEncoded(req *http.Request, sentence string) *http.Request {
	q := req.URL.Query()
	q.Add("q", sentence)

	req.URL.RawQuery = q.Encode()

	return req
}

func httpGet(sentence string) string {
	var result string = "ERROR"

	req := buildRequest("GET", "http://192.168.10.108:3013", "/simplesegment")
	req = queryEncoded(req, sentence)

	response, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		result = gjson.Get(string(body), "segmentresult").String()
	}

	return result
}

func dispatcher(numOfWorkers int, jobs chan string, results chan string) {
	var workers []chan struct{} = make([]chan struct{}, numOfWorkers)

	// running workers
	for i := 0; i < numOfWorkers; i++ {
		workers[i] = worker(jobs, results)
	}

	// wait for workers finished
	for i := 0; i < numOfWorkers; i++ {
		<-workers[i]
		fmt.Printf("Worker %d finished\n", i)
	}

	close(results)
}

func worker(jobs chan string, results chan string) chan struct{} {
	var end chan struct{} = make(chan struct{}, 1)
	go func() {
		for true {
			job, ok := <-jobs
			if !ok {
				break
			}

			results <- httpGet(job)
		}
		end <- struct{}{}
	}()

	return end
}

func tokenize(sentences []string, numOfWorkers int) chan string {
	count := len(sentences)

	var jobs chan string = make(chan string, count)
	var results chan string = make(chan string, count)

	for _, s := range sentences {
		jobs <- s
	}
	close(jobs)

	fmt.Println("START TOKENIZING")
	start_t := time.Now()
	dispatcher(numOfWorkers, jobs, results)
	end_t := time.Now()

	fmt.Println("Tokenize for", len(results), "sentences takes", end_t.Sub(start_t))
	return results
}
