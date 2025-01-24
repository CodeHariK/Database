package trie

import (
	"fmt"
	"sync"
)

type Trie struct {
	root *TrieNode

	sync.RWMutex
}

type TrieNode struct {
	nodes   map[rune]*TrieNode
	present bool

	sync.RWMutex
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		nodes: make(map[rune]*TrieNode),
	}
}

func NewTrie() *Trie {
	return &Trie{
		root: NewTrieNode(),
	}
}

func (t *Trie) PrintAll() {
	var collect func(node *TrieNode, prefix string)
	collect = func(node *TrieNode, prefix string) {
		// If the current node marks the end of a word, print the accumulated prefix
		if node.present {
			fmt.Println(prefix)
		}

		// Recursively visit all child nodes
		for char, child := range node.nodes {
			collect(child, prefix+string(char))
		}
	}

	fmt.Println("+++")
	collect(t.root, "")
	fmt.Println("---")
}

func (t *Trie) Insert(s string) {
	t.Lock()
	defer t.Unlock()

	node := t.root
	for _, c := range s {
		node.Lock()
		if _, ok := node.nodes[c]; !ok {
			node.nodes[c] = NewTrieNode()
		}
		next := node.nodes[c]
		node.Unlock()
		node = next
	}
	node.Lock()
	node.present = true
	node.Unlock()
}

func (t *Trie) Search(s string, needed bool) bool {
	t.Lock()
	defer t.Unlock()

	node := t.root
	for _, char := range s {
		node.RLock()
		next, exists := node.nodes[char]
		node.RUnlock()
		if !exists {
			return false
		}
		node = next
	}
	node.RLock()
	defer node.RUnlock()
	return node.present || !needed
}

func (n *TrieNode) Delete(s string) bool {
	n.Lock()
	defer n.Unlock()

	if len(s) == 0 {
		if n.present {
			n.present = false
			return len(n.nodes) == 0
		}
		return false
	}

	c := rune(s[0])
	if child, ok := n.nodes[c]; ok && child.Delete(s[1:]) {
		delete(n.nodes, c)
		return len(n.nodes) == 0
	}
	return false
}

func (t *Trie) Delete(s string) bool {
	return t.root.Delete(s)
}

func (t *Trie) WordsWithPrefix(prefix string) []string {
	t.RLock()
	defer t.RUnlock()

	node := t.root
	for _, char := range prefix {
		node.RLock()
		next, exists := node.nodes[char]
		node.RUnlock()
		if !exists {
			return nil
		}
		node = next
	}

	return node.wordsWithPrefix(prefix)
}

func (n *TrieNode) wordsWithPrefix(prefix string) []string {
	var results []string
	var dfs func(node *TrieNode, currentPrefix string)
	dfs = func(node *TrieNode, currentPrefix string) {
		node.RLock()
		defer node.RUnlock()
		if node.present {
			results = append(results, currentPrefix)
		}
		for char, child := range node.nodes {
			dfs(child, currentPrefix+string(char))
		}
	}
	dfs(n, prefix)
	return results
}
