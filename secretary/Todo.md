* ```go
allChildren := make([]int32, numNodes*childrenSize) // One big allocation

for i := range nodes {
    nodes[i].children = allChildren[i*childrenSize : (i+1)*childrenSize] // No new allocation
}
```
* Delete key, if deletes node, keep deleted node in array for removal from disk
* Images, binary data visual
* Kademlia
* Persist to storage, with compression
* Bufferpool, Timebaseminheap
* Inverted tree, index, ngram, bm25
* hyperloglog, bloomfilter
* Wal
* Transaction concurrency
