package algorithm

import (
	"sync"
)

// ============================================================================
// 10. 字典树 (Trie) - 前缀搜索优化
// ============================================================================

// TrieNode 字典树节点
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	value    interface{}
}

// Trie 字典树
type Trie struct {
	root *TrieNode
	mu   sync.RWMutex
}

// NewTrie 创建字典树
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

// Insert 插入单词
func (t *Trie) Insert(word string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, ch := range word {
		if node.children[ch] == nil {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.value = value
}

// Search 精确搜索
func (t *Trie) Search(word string) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, ch := range word {
		if node.children[ch] == nil {
			return nil, false
		}
		node = node.children[ch]
	}
	return node.value, node.isEnd
}

// StartsWith 前缀搜索
// 应用: 商品名称自动完成、搜索建议
func (t *Trie) StartsWith(prefix string) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, ch := range prefix {
		if node.children[ch] == nil {
			return nil
		}
		node = node.children[ch]
	}

	results := make([]interface{}, 0)
	t.dfs(node, &results)
	return results
}

// dfs 深度优先搜索收集结果
func (t *Trie) dfs(node *TrieNode, results *[]interface{}) {
	if node.isEnd {
		*results = append(*results, node.value)
	}

	for _, child := range node.children {
		t.dfs(child, results)
	}
}
