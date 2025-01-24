package trie

import (
	"reflect"
	"sort"
	"testing"
)

func TestRuneTrieSearchByPrefix(t *testing.T) {
	trie := NewTrie()

	trie.Insert("c")
	trie.Insert("apple")
	trie.Insert("banana")
	trie.Insert("alphabet")
	trie.Insert("alcohol")

	actual := trie.WordsWithPrefix("a")
	sort.Strings(actual)

	expected := []string{"alcohol", "alphabet", "apple"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%v != %v", actual, expected)
	}
}

func TestTrieSearch(t *testing.T) {
	// Initialize a new Trie
	trie := NewTrie()

	// Insert words
	words := []string{"cat", "car", "cart", "dog", "dove"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Define test cases
	tests := []struct {
		name     string
		input    string
		needed   bool
		expected bool
	}{
		{"Exact match - present", "cat", true, true},
		{"Exact match - absent", "bat", true, false},
		{"Prefix match - present", "car", true, true},
		{"Prefix search without exact match", "ca", false, true},
		{"Unrelated word", "rat", true, false},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := trie.Search(tc.input, tc.needed)
			if result != tc.expected {
				t.Errorf("Search(%q, %t) = %t; want %t", tc.input, tc.needed, result, tc.expected)
			}
		})
	}
}

func TestTrieDelete(t *testing.T) {
	// Initialize a new Trie
	trie := NewTrie()

	// Insert words
	words := []string{"cat", "car", "cart", "dog"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Define test cases
	tests := []struct {
		name          string
		input         string
		expectedFound bool // Expected result of Search after deletion
	}{
		{"Delete existing word", "cat", false},
		{"Delete prefix but keep longer word", "car", false}, // "cart" should remain
		{"Delete longer word", "cart", false},
		{"Delete non-existent word", "bat", false},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Perform deletion
			trie.Delete(tc.input)

			// Check if the word is still found in the Trie
			found := trie.Search(tc.input, true)
			if found != tc.expectedFound {
				t.Errorf("After Delete(%q), Search(%q, true) = %t; want %t", tc.input, tc.input, found, tc.expectedFound)
			}
		})
	}
}
