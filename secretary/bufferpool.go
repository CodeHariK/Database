package secretary

/*
Buffer Pool in Databases

A buffer pool is a memory management component in a database system that caches frequently accessed data pages in RAM. It helps reduce disk I/O by keeping recently used or frequently needed data in memory, improving query performance.

How It Works:
	1.	When a query needs a page from the database, the database engine first checks the buffer pool.
	2.	If the page is in memory (cache hit), it is retrieved quickly.
	3.	If the page is not in memory (cache miss), it is read from disk and placed in the buffer pool.
	4.	If the buffer pool is full, an existing page is evicted using a replacement policy (e.g., LRU - Least Recently Used).
	5.	Modified pages (dirty pages) are periodically written back to disk (checkpointing or background flushing).

Buffer Pool Advantages:
	•	Minimizes disk I/O by keeping frequently accessed pages in RAM.
	•	Speeds up query execution by reducing the need for slow disk reads.
	•	Manages concurrency efficiently by allowing multiple transactions to work on cached pages.
*/
