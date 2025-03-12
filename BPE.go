package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Steps:
// 1. Create a mapping from the chars to token_ids (Probably will use 256 ASCII codes)
// 2. Convert the text to the token_ids format
// 3. Count consecutive pairs and return the topmost pair
// 4. Replace the position of token pairs with the new token_id
// 5. Update the mapping of this new token_ids

// Mapping Class
type Mapping struct {
	vocab        map[string]int
	inverseVocab map[int]string
}

// Source: https://raw.githubusercontent.com/karpathy/char-rnn/master/data/tinyshakespeare/input.txt
func readFile(filename string) (string, error) {
	// Function to read the input file and store in buffer
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// This creates the initial mapping table from ASCII 0-->255
func (m *Mapping) initializeMap() {
	m.vocab = make(map[string]int)
	m.inverseVocab = make(map[int]string)
	for i := 0; i < 256; i++ {
		char := string(rune(i))
		m.vocab[char] = i
		m.inverseVocab[i] = char
	}
}

func (m *Mapping) updateMap(tokens []int) []int {
	freqTable := make(map[[2]int]int)

	// To track the pair with max frequency
	var maxPair [2]int
	maxFreq := 0

	//  To calculate bytePair frequencies
	for i := 0; i < len(tokens)-1; i++ {
		bytePair := [2]int{tokens[i], tokens[i+1]}
		freqTable[bytePair]++

		// Update maxPair if new max freq is found
		if freqTable[bytePair] > maxFreq {
			maxPair = bytePair
			maxFreq = freqTable[bytePair]
		}
	}

	// Creates a new token as the len of the vocab mapping
	newToken_id := len(m.vocab)
	newVocab := m.tokenToText(maxPair[:], false)

	m.vocab[newVocab] = newToken_id
	m.inverseVocab[newToken_id] = newVocab

	return mergeToken(tokens, maxPair, newToken_id)

}

// Merge the two token_id pairs with max frequency into a new tokenList
func mergeToken(tokenList []int, oldBytePair [2]int, newTokenID int) []int {
	token := []int{}
	for i := 0; i < len(tokenList); i++ {
		if i != len(tokenList)-1 && oldBytePair == [2]int{tokenList[i], tokenList[i+1]} {
			token = append(token, newTokenID)
			i++
		} else {
			token = append(token, tokenList[i])
		}
	}
	return token
}

func (m *Mapping) textToToken(inputText string) []int {
	// Create a slice with len 0 and capacity of text
	tokens := make([]int, 0, len(inputText))

	for _, char := range inputText {
		tokens = append(tokens, m.vocab[string(char)])
	}

	return tokens
}

func (m *Mapping) tokenToText(tokens []int, print bool) string {

	result := ""
	for _, token := range tokens {
		result += m.inverseVocab[token]

		if print {
			// Formats in the form
			// token ---> string
			fmt.Printf("%d --->%s\n", token, m.inverseVocab[token])
		}

	}
	return result
}

// Train the BPE model from a given dataset
func train(filename string, maps *Mapping) {
	maps.initializeMap()
	text, err := readFile(filename)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	text = text[:5000]
	// print(text)
	tokens := maps.textToToken(text)
	vocabSize := 1256

	for len(maps.vocab) < vocabSize {
		tokens = maps.updateMap(tokens)
	}
	fmt.Printf("\nFinal vocabulary size: %d\n", len(maps.vocab))

}

func main() {
	var input string
	fmt.Printf("Enter the text to encode: ")

	// Basic code to read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}

	fileName := "input.txt"

	trainedMap := Mapping{}
	train(fileName, &trainedMap)

	tokens := trainedMap.textToToken(input)
	fmt.Println(tokens)
	fmt.Print(trainedMap.tokenToText(tokens, true))
	// To do: Implement reverse searching in the mappings

}
