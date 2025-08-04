package indexer

import (
	"fmt"
	"strings"
)

type wordKey string
type occurenceFrequencyType int
type document string
type charPositionIndicesType []int
type WordIndex struct {
	DocumentName  string
	CharPositions charPositionIndicesType
}
type DocumentIndex struct {
	Name                     string
	WordsReverseIndexByCount map[wordKey]map[occurenceFrequencyType][]WordIndex
	Documents                map[string]document
}

func CreateReverseIndex(indexName string) DocumentIndex {
	return DocumentIndex{
		Name:                     indexName,
		WordsReverseIndexByCount: make(map[wordKey]map[occurenceFrequencyType][]WordIndex),
		Documents:                make(map[string]document),
	}
}
func (docIndex DocumentIndex) LoadDocument(docName, doc string) error {
	var charIndex int = 0
	wordsReverseIndexByDocument := make(map[wordKey]map[string]charPositionIndicesType)
	for _, word := range strings.Fields(doc) {
		newCharIndex := charIndex + (1 + len(word))
		word := wordKey(strings.Trim(word, ".,:;'\"()[]{}|\\/"))
		if wordsReverseIndexByDocument[word] == nil {
			wordsReverseIndexByDocument[word] = make(map[string]charPositionIndicesType)
		}
		wordsReverseIndexByDocument[word][docName] = append(wordsReverseIndexByDocument[word][docName], charIndex)
		charIndex = newCharIndex
	}
	for word := range wordsReverseIndexByDocument {
		for docName, indices := range wordsReverseIndexByDocument[word] {
			newWordIndex := WordIndex{DocumentName: docName, CharPositions: indices}
			occurenceFrequency := occurenceFrequencyType(len(newWordIndex.CharPositions))
			if docIndex.WordsReverseIndexByCount[word] == nil {
				docIndex.WordsReverseIndexByCount[word] = make(map[occurenceFrequencyType][]WordIndex)
			}
			docIndex.WordsReverseIndexByCount[word][occurenceFrequency] = append(docIndex.WordsReverseIndexByCount[word][occurenceFrequency], newWordIndex)
		}
	}
	docIndex.Documents[docName] = document(doc)
	return nil
}
func (docIndex DocumentIndex) ListDocuments() ([]string, error) {
	var returnString []string
	for documentName := range docIndex.Documents {
		returnString = append(returnString, documentName)
	}
	return returnString, nil
}
func (docIndex DocumentIndex) FindWord(word string) (map[occurenceFrequencyType][]WordIndex, error) {
	index, doesExist := docIndex.WordsReverseIndexByCount[wordKey(word)]
	if !doesExist {
		return nil, fmt.Errorf("word %s have not yet been indexed", word)
	}
	return index, nil
}
func (docIndex DocumentIndex) DumpDocuments() map[string]document {
	return docIndex.Documents
}
