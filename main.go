package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type imageURLID struct {
	url string
	id  string
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	images, err := fetchImageURLIDs()
	if err != nil {
		fmt.Println("Could not read the ImageURLIDs from file")
		os.Exit(1)
	}
	loadImages(images)
}
func loadImages(images []imageURLID) {

	//channel for terminating the workers
	killsignal := make(chan bool)

	//queue of jobs
	q := make(chan imageURLID)

	// done channel takes the result of the job
	done := make(chan bool)

	numberOfWorkers := 10
	for i := 0; i < numberOfWorkers; i++ {
		go worker(q, i, done, killsignal)
	}

	numberOfJobs := len(images)

	for _, image := range images {
		go func(image imageURLID) {
			q <- image
		}(image)
	}

	for c := 0; c < numberOfJobs; c++ {
		<-done
	}

	fmt.Println("finished")

	// cleaning workers
	close(killsignal)
	time.Sleep(2 * time.Second)

}

func downloadImage(image imageURLID) {
	response, e := http.Get(image.url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()
	imageBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		//TODO: proper error reporting
		fmt.Println(err)
	}
	saveBytesToImage(imageBytes, image.id)
}

func saveBytesToImage(imageBytes []byte, id string) {
	err := ioutil.WriteFile(path.Join("images", strings.Join([]string{id, ".jpg"}, "")), imageBytes, 0644)
	if err != nil {
		// error handling
		fmt.Println(err)
	}
}

func fetchImageURLIDs() ([]imageURLID, error) {
	imageIDs, err := readLines("imageIDs.txt")
	if err != nil {
		return nil, err
	}
	imageURLs, err := readLines("imageURLs.txt")
	if err != nil {
		return nil, err
	}
	imageURLIDs := make([]imageURLID, len(imageIDs))
	for i, line := range imageIDs {
		imageURLIDs[i].id = line
		imageURLIDs[i].url = imageURLs[i]
	}
	return imageURLIDs, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func worker(queue chan imageURLID, worknumber int, done, ks chan bool) {
	for true {
		select {
		case k := <-queue:
			fmt.Println("doing work!", k, "worknumber", worknumber)
			downloadImage(k)
			done <- true
		case <-ks:
			fmt.Println("worker halted, number", worknumber)
			return
		}
	}
}
