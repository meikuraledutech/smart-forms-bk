package buffer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ResponseData holds a complete response submission
type ResponseData struct {
	ResponseID      string
	FormID          string
	TotalTimeSpent  int
	FlowPath        []string
	Metadata        map[string]interface{}
	Answers         []AnswerData
}

// AnswerData holds answer information
type AnswerData struct {
	ResponseID       string
	FlowConnectionID string
	AnswerText       string
	AnswerValue      map[string]interface{}
	TimeSpent        *int
}

// ResponseBuffer handles buffered batch inserts
type ResponseBuffer struct {
	db          *pgxpool.Pool
	queue       chan ResponseData
	batchSize   int
	flushTicker *time.Ticker
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewResponseBuffer creates a new response buffer
func NewResponseBuffer(db *pgxpool.Pool, queueSize, batchSize int, flushInterval time.Duration) *ResponseBuffer {
	ctx, cancel := context.WithCancel(context.Background())

	rb := &ResponseBuffer{
		db:          db,
		queue:       make(chan ResponseData, queueSize),
		batchSize:   batchSize,
		flushTicker: time.NewTicker(flushInterval),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start worker goroutine
	go rb.worker()

	log.Printf("Response buffer initialized (queue: %d, batch: %d, flush: %v)",
		queueSize, batchSize, flushInterval)

	return rb
}

// Enqueue adds a response to the buffer
func (rb *ResponseBuffer) Enqueue(data ResponseData) error {
	select {
	case rb.queue <- data:
		return nil
	case <-rb.ctx.Done():
		return rb.ctx.Err()
	default:
		// Queue is full, fallback to immediate insert
		log.Println("Warning: Response buffer full, inserting immediately")
		return rb.insertBatch([]ResponseData{data})
	}
}

// worker processes the buffer
func (rb *ResponseBuffer) worker() {
	batch := make([]ResponseData, 0, rb.batchSize)

	for {
		select {
		case data := <-rb.queue:
			// Add to batch
			batch = append(batch, data)

			// TRIGGER 1: Buffer full
			if len(batch) >= rb.batchSize {
				if err := rb.insertBatch(batch); err != nil {
					log.Printf("Error inserting batch: %v", err)
				}
				batch = batch[:0] // Clear batch
			}

		case <-rb.flushTicker.C:
			// TRIGGER 2: Timer expired
			if len(batch) > 0 {
				if err := rb.insertBatch(batch); err != nil {
					log.Printf("Error flushing batch: %v", err)
				}
				batch = batch[:0] // Clear batch
			}

		case <-rb.ctx.Done():
			// Shutdown: flush remaining
			if len(batch) > 0 {
				if err := rb.insertBatch(batch); err != nil {
					log.Printf("Error flushing final batch: %v", err)
				}
			}
			return
		}
	}
}

// insertBatch performs batch insert
func (rb *ResponseBuffer) insertBatch(batch []ResponseData) error {
	if len(batch) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start transaction
	tx, err := rb.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Batch insert responses
	for _, data := range batch {
		flowPathJSON, _ := json.Marshal(data.FlowPath)
		metadataJSON, _ := json.Marshal(data.Metadata)

		_, err := tx.Exec(ctx, `
			INSERT INTO form_responses (id, form_id, total_time_spent, flow_path, metadata)
			VALUES ($1, $2, $3, $4, $5)
		`, data.ResponseID, data.FormID, data.TotalTimeSpent, flowPathJSON, metadataJSON)
		if err != nil {
			return err
		}

		// Batch insert answers for this response
		for _, answer := range data.Answers {
			var answerValueJSON []byte
			if answer.AnswerValue != nil {
				answerValueJSON, _ = json.Marshal(answer.AnswerValue)
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO response_answers (response_id, flow_connection_id, answer_text, answer_value, time_spent)
				VALUES ($1, $2, $3, $4, $5)
			`, answer.ResponseID, answer.FlowConnectionID, answer.AnswerText, answerValueJSON, answer.TimeSpent)
			if err != nil {
				return err
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Printf("Batch inserted %d responses successfully", len(batch))
	return nil
}

// Close shuts down the buffer gracefully
func (rb *ResponseBuffer) Close() {
	rb.flushTicker.Stop()
	rb.cancel()
	close(rb.queue)
	log.Println("Response buffer closed")
}
