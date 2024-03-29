package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var inputFileName = flag.String("csv", "problems.csv",
	"a csv file in the format of 'question,answer'")

var limit = flag.Int("limit", 30,
	"the time limit for the quiz in seconds")

var shuffle = flag.Bool("shuffle", false, "shuffle questions")

type question struct {
	question, answer string
}

func main() {
	flag.Parse()

	reader := makeCsvReader(*inputFileName)
	questions := parseQuestions(reader)

	if *shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(questions), func(i, j int) {
			questions[i], questions[j] = questions[j], questions[i]
		})
	}

	inputReader := bufio.NewReader(os.Stdin)
	fmt.Print("Press a key to start quiz")
	inputReader.ReadRune()
	correctChan, endChan := startQuiz(questions)
	timerChan := time.After(time.Duration(*limit) * time.Second)

	numCorrect := 0

MainLoop:
	for {
		select {
		case <-correctChan:
			numCorrect++
		case <-endChan:
			break MainLoop
		case <-timerChan:
			fmt.Println()
			break MainLoop
		}
	}

	fmt.Printf("You scored %d out of %d.\n", numCorrect, len(questions))
}

func makeCsvReader(filename string) io.Reader {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("unable to open '%s': %s\n", filename, err)
		return nil
	}

	return bufio.NewReader(file)
}

func parseQuestions(r io.Reader) []question {
	csvReader := csv.NewReader(r)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("error reading csv file: %s\n", err)
		return nil
	}

	questions := make([]question, 0, len(records))
	for i, record := range records {
		if len(record) != 2 {
			log.Printf("incorrect number of fields on line %d", i+1)
			continue
		}

		q, a := record[0], strings.ToLower(strings.TrimSpace(record[1]))
		questions = append(questions, question{q, a})
	}

	return questions
}

func startQuiz(questions []question) (<-chan bool, <-chan bool) {
	correctChan, endChan := make(chan bool, 0), make(chan bool, 0)
	go func(questions []question) {
		scanner := bufio.NewScanner(os.Stdin)
		for i, q := range questions {
			fmt.Printf("Problem #%d: %s = ", i+1, q.question)
			ok := scanner.Scan()
			err := scanner.Err()
			if !ok && err != nil {
				log.Fatalf("error reading from input: %s", err)
			}

			answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if answer == q.answer {
				correctChan <- true
			}
		}

		endChan <- true
		close(correctChan)
		close(endChan)
	}(questions)

	return correctChan, endChan
}
