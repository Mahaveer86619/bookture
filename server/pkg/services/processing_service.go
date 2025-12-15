package services

import (
	"log"
	"sync"
)

type Job func(reportProgress func(int))

type ProcessingService struct {
	jobQueue    chan jobRequest
	progressMap sync.Map
	wg          sync.WaitGroup
	quit        chan bool
}

type jobRequest struct {
	id  string
	job Job
}

func NewProcessingService(workerCount int, queueSize int) *ProcessingService {
	ps := &ProcessingService{
		jobQueue: make(chan jobRequest, queueSize),
		quit:     make(chan bool),
	}

	// Start workers
	for i := range workerCount {
		ps.wg.Add(1)
		go ps.worker(i)
	}

	return ps
}

func (ps *ProcessingService) worker(id int) {
	defer ps.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case req := <-ps.jobQueue:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Worker %d panicked: %v", id, r)
						ps.progressMap.Store(req.id, -1) // Error state
					}
				}()

				// Initialize progress
				ps.progressMap.Store(req.id, 0)

				// Run job with a reporter function
				req.job(func(p int) {
					ps.progressMap.Store(req.id, p)
				})

				// Cleanup or mark complete
				ps.progressMap.Store(req.id, 100)
			}()
		case <-ps.quit:
			return
		}
	}
}

// Enqueue adds a job to the queue. Non-blocking; drops job if queue is full.
func (ps *ProcessingService) Enqueue(id string, job Job) {
	select {
	case ps.jobQueue <- jobRequest{id: id, job: job}:
	default:
		log.Println("Warning: Job queue is full, job dropped")
	}
}

// GetProgress is the handler to get update percentage
func (ps *ProcessingService) GetProgress(id string) int {
	if val, ok := ps.progressMap.Load(id); ok {
		return val.(int)
	}
	return 0
}

func (ps *ProcessingService) Shutdown() {
	close(ps.quit)
	ps.wg.Wait()
}
