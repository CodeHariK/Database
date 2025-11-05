
#### Actor Model	
* Actors are independent entities that communicate via messages.
* State Encapsulated in actors (no shared memory).
* Communication Asynchronous message passing.
* Synchronization	No explicit locks; actors process messages sequentially.
* Deadlocks Avoided by design.
* Scalability High; actors can be distributed across nodes.
* Complexity Moderate; requires message design and failure handling.
* Best for distributed systems, fault tolerance, and large-scale concurrent applications.

#### CSP (Communicating Sequential Processes)	
* Processes (goroutines, threads) communicate via channels.
* Can have shared memory but encourages message passing.
* Synchronous or buffered channels.
* Blocking or non-blocking channel operations.
* Deadlocks Possible but less likely if channels are used correctly.
* Scalability High; well-suited for many-core systems.
* Complexity Moderate; channel communication must be structured properly.
* Best for structured concurrency, efficient CPU utilization, and lightweight concurrency management.

#### Locks (Mutexes, RWLocks, etc.)
* Threads or goroutines synchronize access to shared memory.
* Shared memory with explicit synchronization.
* Direct access to shared memory, requires locks.
* Mutexes, RW locks, and atomic operations.
* Deadlocks Common issue if locks are misused.
* Scalability Low; locks can cause contention.
* Complexity High; requires careful lock management.
* Best for shared memory, low-latency applications, and fine-grained synchronization.
