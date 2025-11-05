package main

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

// ================= Actor System =================

// Message struct with priority (lower value = higher priority)
type Message struct {
	sender   string
	content  string
	priority int
}

// MessageQueue implements a priority queue
type MessageQueue []*Message

func (mq MessageQueue) Len() int { return len(mq) }
func (mq MessageQueue) Less(i, j int) bool {
	return mq[i].priority < mq[j].priority // Lower priority number = higher priority
}
func (mq MessageQueue) Swap(i, j int) { mq[i], mq[j] = mq[j], mq[i] }

func (mq *MessageQueue) Push(x interface{}) {
	*mq = append(*mq, x.(*Message))
}

func (mq *MessageQueue) Pop() interface{} {
	old := *mq
	n := len(old)
	item := old[n-1]
	*mq = old[0 : n-1]
	return item
}

// Actor struct
type Actor struct {
	name      string
	mailbox   MessageQueue
	mailboxMu sync.Mutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// NewActor creates a new actor with a priority queue
func NewActor(name string) *Actor {
	a := &Actor{
		name:     name,
		mailbox:  MessageQueue{},
		stopChan: make(chan struct{}),
	}
	heap.Init(&a.mailbox)
	go a.start()
	return a
}

// start begins processing messages
func (a *Actor) start() {
	log.Printf("[%s] started\n", a.name)
	for {
		a.mailboxMu.Lock()
		if len(a.mailbox) == 0 {
			a.mailboxMu.Unlock()
			select {
			case <-a.stopChan:
				log.Printf("[%s] shutting down\n", a.name)
				return
			default:
				time.Sleep(100 * time.Millisecond) // Avoid busy-waiting
			}
			continue
		}

		// Process the highest-priority message
		msg := heap.Pop(&a.mailbox).(*Message)
		a.mailboxMu.Unlock()

		a.processMessage(msg)
	}
}

// processMessage simulates message handling with error handling and retries
func (a *Actor) processMessage(msg *Message) {
	log.Printf("[%s] received message from [%s]: %s (Priority: %d)\n", a.name, msg.sender, msg.content, msg.priority)

	// Simulate failure with a 30% chance
	if time.Now().UnixNano()%3 == 0 {
		log.Printf("[%s] failed to process message, retrying...\n", a.name)
		time.Sleep(200 * time.Millisecond)
		a.SendMessage(msg.sender, msg.content, msg.priority+1) // Increase priority for retry
		return
	}

	time.Sleep(500 * time.Millisecond) // Simulate processing time
}

// SendMessage adds a message to the actor's priority queue
func (a *Actor) SendMessage(sender, content string, priority int) {
	a.mailboxMu.Lock()
	heap.Push(&a.mailbox, &Message{sender: sender, content: content, priority: priority})
	a.mailboxMu.Unlock()
}

// Stop stops the actor
func (a *Actor) Stop() {
	close(a.stopChan)
}

// ================= Actor Registry =================

// Global registry to track actors
var (
	actorRegistry   = make(map[string]*Actor)
	registryLock    sync.Mutex
	supervisorActor *Actor
)

// RegisterActor adds an actor to the registry
func RegisterActor(actor *Actor) {
	registryLock.Lock()
	defer registryLock.Unlock()
	actorRegistry[actor.name] = actor
}

// GetActor retrieves an actor from the registry
func GetActor(name string) *Actor {
	registryLock.Lock()
	defer registryLock.Unlock()
	return actorRegistry[name]
}

// ================= Supervisor Actor =================

// Supervisor monitors actors and restarts them on failure
func Supervisor() {
	supervisorActor = NewActor("Supervisor")
	go func() {
		for {
			time.Sleep(1 * time.Second)
			log.Println("[Supervisor] Monitoring actors...")

			registryLock.Lock()
			for name, actor := range actorRegistry {
				select {
				case <-actor.stopChan:
					log.Printf("[Supervisor] Restarting actor: %s\n", name)
					newActor := NewActor(name)
					RegisterActor(newActor)
				default:
				}
			}
			registryLock.Unlock()
		}
	}()
}

// ================= Main =================
func main() {
	log.Println("Starting Actor System...")

	// Start Supervisor
	Supervisor()

	// Create actors
	actorA := NewActor("ActorA")
	actorB := NewActor("ActorB")
	RegisterActor(actorA)
	RegisterActor(actorB)

	// Send messages with different priorities
	actorA.SendMessage("Main", "Hello ActorA!", 2)
	actorB.SendMessage("Main", "Hello ActorB!", 1)
	actorA.SendMessage("ActorB", "Urgent message!", 0)
	actorB.SendMessage("ActorA", "Regular message.", 3)

	// Simulate running for some time
	time.Sleep(5 * time.Second)

	// Stop actors
	actorA.Stop()
	actorB.Stop()

	log.Println("Actor System shutting down.")
}
