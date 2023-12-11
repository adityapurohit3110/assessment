package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestBody struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponseBody struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func sortSequential(arrays [][]int) [][]int {
	sortedArrays := make([][]int, len(arrays))

	for i, subArray := range arrays {
		sortedSubArray := make([]int, len(subArray))
		copy(sortedSubArray, subArray)
		sort.Ints(sortedSubArray)
		sortedArrays[i] = sortedSubArray
	}

	return sortedArrays
}

func sortConcurrent(arrays [][]int) [][]int {
	var wg sync.WaitGroup
	wg.Add(len(arrays))

	sortedArrays := make([][]int, len(arrays))
	for i, subArray := range arrays {
		go func(i int, subArray []int) {
			defer wg.Done()
			sortedSubArray := make([]int, len(subArray))
			copy(sortedSubArray, subArray)
			sort.Ints(sortedSubArray)
			sortedArrays[i] = sortedSubArray
		}(i, subArray)
	}

	wg.Wait()
	return sortedArrays
}

func processHandler(w http.ResponseWriter, r *http.Request, sortFunc func([][]int) [][]int) {
	var requestBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Not a valid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortFunc(requestBody.ToSort)
	timeTaken := time.Since(startTime).Nanoseconds()

	responseBody := ResponseBody{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		http.Error(w, "Not able to encode JSON response", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/process-single", func(w http.ResponseWriter, r *http.Request) {
		processHandler(w, r, sortSequential)
	})

	http.HandleFunc("/process-concurrent", func(w http.ResponseWriter, r *http.Request) {
		processHandler(w, r, sortConcurrent)
	})

	port := 8000
	fmt.Printf("Server is running on :%d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
