#

* [#08 - Tree Indexes: B+Trees (CMU Intro to Database Systems)](https://www.youtube.com/watch?v=scUtG_6M_lU)
* [SQLite: How it works](https://www.youtube.com/watch?v=ZSKLA81tBis)
* [Write a database from scratch](https://www.youtube.com/playlist?list=PLWRwj01AnyEtjaw-ZnnAQWnVYPZF5WayV)

* [Understanding B-Trees: The Data Structure Behind Modern Databases](https://www.youtube.com/watch?v=K1a2Bk8NrYQ)

* [Build a NoSQL Database From Scratch in 1000 Lines of Code](https://medium.com/better-programming/build-a-nosql-database-from-the-scratch-in-1000-lines-of-code-8ed1c15ed924)
* [Writing a SQL database from scratch in Go: 1. SELECT, INSERT, CREATE and a REPL](https://notes.eatonphil.com/database-basics.html)

* [Database Engine Development](https://www.youtube.com/playlist?list=PLm7R-cUo29CXVu9a9TzBEwSQ9JPVGmISg)

* https://github.com/cmu-db/bustub

##

```
1. Persistence. How not to lose or corrupt your data. Recovering from a crash.
2. Indexing. Efficiently querying and manipulating your data. (B-tree).
3. Concurrency. How to handle multiple (large number of ) clients. And transactions.
```

## Vector Db

* https://github.com/skyzh/write-you-a-vector-db

## Search

* [The Art of Searching](https://www.youtube.com/watch?v=yst6VQ7Lwpo)
* [Algorithms & data-structures that power Lucene & ElasticSearch](https://www.youtube.com/watch?v=eQ-rXP-D80U)

* [How do Spell Checkers work? Levenshtein Edit Distance](https://www.youtube.com/watch?v=Cu7Tl7FGigQ)
* [The Algorithm Behind Spell Checkers](https://www.youtube.com/watch?v=d-Eq6x1yssU)

## CRDT




https://github.com/nictuku/dht
https://github.com/shiyanhui/dht




* [Collaborative Text Editing Paper](https://arxiv.org/pdf/2305.00583)

* https://github.com/josephg/crdt-from-scratch
* https://github.com/josephg/egwalker-from-scratch

* [Collaborative Text Editing with Eg-Walker](https://www.youtube.com/watch?v=rjbEG7COj7o)
* [Text CRDTs from scratch, in code!](https://www.youtube.com/watch?v=_lQ2Q4Kzi1I)
* [Lets write Eg-walker from scratch! Part 1](https://www.youtube.com/watch?v=ggXka5TTsOs)

* [Conflict-Free Replicated Data Types (CRDT) for Distributed JavaScript Apps](https://www.youtube.com/watch?v=M8-WFTjZoA0)
* [CRDTs: The Hard Parts](https://www.youtube.com/watch?v=x7drE24geUw)
* [CRDTs and the Quest for Distributed Consistency](https://www.youtube.com/watch?v=B5NULPSiOGw)
* [A CRDT Primer: Defanging Order Theory](https://www.youtube.com/watch?v=OOlnp2bZVRs)

* [Loro Is Local-First State With CRDT](https://www.youtube.com/watch?v=NB7HRfyufLk)
* [How Yjs works from the inside out](https://www.youtube.com/watch?v=0l5XgnQ6rB4)

* [CRDTs for Non Academics](https://www.youtube.com/watch?v=vBU70EjwGfw)
* [An introduction to Conflict-Free Replicated Data Types (CRDTs)](https://www.youtube.com/watch?v=gZP2VUmH05A)

* [CRDT Survey](https://mattweidner.com/2023/09/26/crdt-survey-1.html)
* [An introduction to state-based CRDTs](https://www.bartoszsypytkowski.com/the-state-of-a-state-based-crdts/)

## Data Structure

* [Heaps, heapsort, and priority queues - Inside code](https://www.youtube.com/watch?v=pLIajuc31qk)
* [Trie data structure - Inside code](https://www.youtube.com/watch?v=qA8l8TAMyig)
* [Compressed trie](https://www.youtube.com/watch?v=qakGXuOW1S8)

## Probablistic Data structures

* [Hello Interview : Bloom Filters, Count-Min Sketch, HyperLogLog](https://www.youtube.com/watch?v=IgyU0iFIoqM)

* [Probablistic data structure lectures](https://www.youtube.com/playlist?list=PL2mpR0RYFQsAR5RyB54FyEE9vUiGtCSZM)

### Bloom filter 

* [Wikipedia](https://en.wikipedia.org/wiki/Bloom_filter)
* [mCoding : Bloom Filters](https://www.youtube.com/watch?v=qZNJTh2NEiU)
* [Number0 : Bloom Filters](https://www.youtube.com/watch?v=eCUm4U3WDpM)
* [ByteByteGo : Bloom Filters](https://www.youtube.com/watch?v=V3pzxngeLqw)
* [Spanning Tree : What Are Bloom Filters?](https://www.youtube.com/watch?v=kfFacplFY4Y)
* [ByteMonk : Bloom Filters](https://www.youtube.com/watch?v=GT0En1dGntY)

Bloom filter is a space-efficient probabilistic data structure, that is used to test whether an element is a member of a set. False positive matches are possible, but false negatives are not - in other words, a query returns either "possibly in set" or "definitely not in set". Elements can be added to the set, but not removed.

A Bloom filter is a representation of a set of _n_ items, where the main requirement is to make membership queries; _i.e._, whether an item is a member of a set.

#### Uses
##### Cache filtering
Content delivery networks deploy web caches around the world to cache and serve web content to users with greater performance and reliability. A key application of Bloom filters is their use in efficiently determining which web objects to store in these web caches. To prevent caching one-hit-wonders, a Bloom filter is used to keep track of all URLs that are accessed by users.
##### Web Crawler

### HyperLogLog

* [Wikipedia](https://en.wikipedia.org/wiki/HyperLogLog)
* [PapersWeLove : HyperLogLog](https://www.youtube.com/watch?v=y3fTaxA8PkU)
* [The Algorithm with the Best Name - HyperLogLog Explained](https://www.youtube.com/watch?v=2PlrMCiUN_s)
* [A problem so hard even Google relies on Random Chance](https://www.youtube.com/watch?v=lJYufx0bfpw)
* [Counting BILLIONS with Just Kilobytes](https://www.youtube.com/watch?v=f69hh3KgFEk)
* https://github.com/tylertreat/BoomFilters/blob/master/hyperloglog.go

HyperLogLog is an algorithm for the count-distinct problem, Probabilistic cardinality estimators.

### Count–min sketch

* [Wikepedia](https://en.wikipedia.org/wiki/Count%E2%80%93min_sketch)
* https://github.com/tylertreat/BoomFilters/blob/master/countmin.go
* [Count-min Sketch](https://www.youtube.com/watch?v=Okdjn7o4q8E)

The goal of the basic version of the count–min sketch is to consume a stream of events, one at a time, and count the frequency of the different types of events in the stream.

### HeavyKeeper TopK

* [Understanding Probabilistic Data Structures](https://www.youtube.com/watch?v=2Dzc7fxA0us)

### TDigest

* [Sketching Data with T Digest](https://www.youtube.com/watch?v=ETUYhEZRtWE)

## Cache

* [TinyLFU: A Highly Efficient Cache Admission Policy](https://dgraph.io/blog/refs/TinyLFU%20-%20A%20Highly%20Efficient%20Cache%20Admission%20Policy.pdf)

* [Caffeine Design of a Modern Cache](https://docs.google.com/presentation/d/1NlDxyXsUG1qlVHMl4vsUUBQfAJ2c2NsFPNPr2qymIBs/edit#slide=id.p)
* [Design of a Modern Cache](https://highscalability.com/design-of-a-modern-cache/)
* [The State of Caching in Go](https://dgraph.io/blog/post/caching-in-go/)
* [Introducing Ristretto: A High-Performance Go Cache](https://dgraph.io/blog/post/introducing-ristretto-high-perf-go-cache/)
* [On Window TinyLFU](https://9vx.org/post/on-window-tinylfu/)

* https://en.wikipedia.org/wiki/Cache_replacement_policies#LRU

* https://github.com/hypermodeinc/ristretto
* https://github.com/dgryski/go-tinylfu

## LSM-Tree

[#04 - Database Storage: Log-Structured Merge Trees & Tuples (CMU Intro to Database Systems)](https://www.youtube.com/watch?v=IHtVWGhG0Xg&t=1372s)

https://github.com/facebook/rocksdb/wiki

https://github.com/krasun/lsmtree
https://github.com/skyzh/mini-lsm

## OLAP

* https://github.com/risinglightdb/risinglight

## Concurrency

* [Golang concurrency - Locks, Lock Free and everything in between](https://www.youtube.com/watch?v=gNQ6j2Y2HFs)
* [Optimistic Locking clearly explained](https://www.youtube.com/watch?v=d41JuPT_Wls)
* [Understanding the Disruptor](https://www.youtube.com/watch?v=DCdGlxBbKU4)

* https://en.wikipedia.org/wiki/Actor_model

* The Art of Multiprocessor Programming
* Seven Concurrency Models in Seven Weeks

* https://proto.actor/docs/

* [The Actor Model](https://www.youtube.com/watch?v=7erJ1DV_Tlo)
* [A brief introduction to the actor model & distributed actors](https://www.youtube.com/watch?v=YTQeJegJnbo)
* [Introduction to the Actor Model for Concurrent Computation](https://www.youtube.com/watch?v=lPTqcecwkJg)

* https://github.com/vladopajic/go-actor

* https://www.brianstorti.com/the-actor-model/

* [Code : Ring Buffer](https://www.youtube.com/watch?v=KyreJSKEagg)
* https://kmdreko.github.io/posts/20191003/a-simple-lock-free-ring-buffer/

* [Producer/Consumer, The RingBuffer and The Log](https://www.youtube.com/watch?v=uqSeuGQhnf0)

* [Disruptor](https://lmax-exchange.github.io/disruptor/files/Disruptor-1.0.pdf)
* https://github.com/LMAX-Exchange/disruptor/wiki/Blogs-And-Articles

* https://github.com/smarty-prototypes/go-disruptor

## Compress

* [Data Encodings used by Columnar and Time series databases](https://www.youtube.com/watch?v=wUO2snhiosk)
