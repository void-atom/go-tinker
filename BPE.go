package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Steps:
// Train
// 1. Create a mapping from the chars to token_ids (Probably will use 256 ASCII codes)
// 2. Convert the text to the token_ids format
// 3. Count consecutive pairs and return the topmost pair
// 4. Replace the position of token pairs with the new token_id
// 5. Update the mapping of this new token_ids

// Inference:
// 1. Sort the vocabs hashMap and convert it into list/slices of strings
// 2. Start matching the longest substring and convert it to tokenized_text(Uses Binary Tree to preserve order)
// 3. InorderTraversal of binary tree to get back the tokenized_text in order
// 4. Use the look-up table(vocabs) to convert to tokenized_text to tokens
// 5. Use another look-up (inverseVocab) to convert it back to input Text

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
		if tokens[i+1] == ' ' {
			continue
		}
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

func (m *Mapping) saveMappings() bool {
	fmt.Println("Creating \"mappings.dump\" in the present directory")

	// Create a file mappings.dump
	file, err := os.Create("mappings.dump")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Save using gob Encoding
	// Normal write to file was not used because vocab contains UTF-8 encodings
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(m.vocab)
	if err != nil {
		panic(err)
	}
	err = encoder.Encode(m.inverseVocab)
	if err != nil {
		panic(err)
	}

	fmt.Println("Data saved in binary format.")
	return true

}

func getMappings(vocab *map[string]int, inverseVocab *map[int]string) {
	fmt.Println("Retrieving mappings from \"mappings.dump\"")
	file, err := os.Open("mappings.dump")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a decoder
	decoder := gob.NewDecoder(file)

	// Decode the first map
	err = decoder.Decode(vocab)
	if err != nil {
		panic(err)
	}

	// Decode the second map
	err = decoder.Decode(inverseVocab)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully retrived the Mappings and inverse Mappings.")

	// Print the results
	// fmt.Print("Decoded Vocab:", vocab)
	// fmt.Print("Decoded InverseVocab:", inverseVocab)
}

// Train the BPE model from a given dataset
func train(filename string, maps *Mapping) {

	// Checking if the file exists already
	_, err := os.Stat("mappings.dump")
	if os.IsNotExist(err) {
		fmt.Println("Mapping file does not exist. Training on \"input.txt\" to generate \"mappings.dump\"")
	} else if err != nil {
		fmt.Println("Error checking file:", err)
	} else {
		fmt.Println("Mapping file exists.")
		return
	}

	maps.initializeMap()
	text, err := readFile(filename)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	tokens := maps.textToToken(text)
	vocabSize := 2356

	for len(maps.vocab) < vocabSize {
		tokens = maps.updateMap(tokens)
	}
	// fmt.Print(maps.vocab)

	fmt.Printf("\nFinal vocabulary size: %d\n", len(maps.vocab))
	maps.saveMappings()
}

func sortMapByKeyLengthDescending(m map[string]int) ([]string, []int) {
	type kv struct {
		Key   string
		Value int
	}
	var sortedSlice []kv

	for k, v := range m {
		sortedSlice = append(sortedSlice, kv{Key: k, Value: v})
	}

	// Sort by key length in descending order
	sort.Slice(sortedSlice, func(i, j int) bool {
		return len(sortedSlice[i].Key) > len(sortedSlice[j].Key)
	})

	var keys []string
	var values []int
	for _, item := range sortedSlice {
		keys = append(keys, item.Key)
		values = append(values, item.Value)
	}

	return keys, values
}

type TreeNode struct {
	text  string
	Left  *TreeNode
	Right *TreeNode
}

func processNode(input string, node *TreeNode, vocabs []string) {
	if len(input) <= 0 {
		return
	}
	index := -1
	var matched string
	for _, word := range vocabs {
		if strings.Index(input, word) >= 0 {
			index = strings.Index(input, word)
			matched = word
			break
		}
	}

	node.text = input[index : index+len(matched)]

	if len(input[:index]) > 0 {
		node.Left = &TreeNode{}
		processNode(input[:index], node.Left, vocabs)
	}

	if len(input[index+len(matched):]) > 0 {
		node.Right = &TreeNode{}
		processNode(input[index+len(matched):], node.Right, vocabs)
	}

}

func InOrderTraversal(root *TreeNode, traversalBucket *[]string) {
	if root != nil {
		InOrderTraversal(root.Left, traversalBucket)
		fmt.Print(root.text, "|")
		*traversalBucket = append(*traversalBucket, root.text)

		InOrderTraversal(root.Right, traversalBucket)
	}
}

func encode(vocab map[string]int, input string) []int {

	sortedVocabs, _ := sortMapByKeyLengthDescending(vocab)
	var tokenized []string

	// Using regex to break words preserving the leading spaces
	re := regexp.MustCompile(`(\s*\S+|\s+)`)
	words := re.FindAllString(input, -1)
	fmt.Print("Tokenized words: ")

	for _, word := range words {
		root := &TreeNode{}
		processNode(word, root, sortedVocabs)
		InOrderTraversal(root, &tokenized)

	}

	println()

	var tokens []int
	for _, tokenWord := range tokenized {
		tokens = append(tokens, vocab[tokenWord])
	}
	return tokens

}

func decode(tokens []int, inverseVocab map[int]string) string {
	decodedText := ""

	for _, token := range tokens {
		decodedText += inverseVocab[token]
	}
	return decodedText
}

func getInput() string {
	println()
	fmt.Printf("Enter the text to encode: ")
	var input string
	// Basic code to read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
	return input
}

func main() {
	fileName := "input.txt"

	trainedMap := Mapping{}
	train(fileName, &trainedMap)

	var invVocabDecoded map[int]string
	var vocabDecoded map[string]int

	getMappings(&vocabDecoded, &invVocabDecoded)

	inputText := "We go to the moon not because it's easy but because it's hard." //getInput()

	// fmt.Println("Input text :", inputText)
	tokenList := encode(vocabDecoded, inputText)
	fmt.Println("Token list: ", tokenList)
	fmt.Println("Decoded text:", decode(tokenList, invVocabDecoded))
}
