package service

import (
	"context"
	"testing"
)

type recordingSyncProducer struct {
	upsertedProblemIDs []uint
	deletedProblemIDs  []uint
}

func (p *recordingSyncProducer) EnqueueProblemUpsert(_ context.Context, problemID uint) error {
	p.upsertedProblemIDs = append(p.upsertedProblemIDs, problemID)
	return nil
}

func (p *recordingSyncProducer) EnqueueProblemDelete(_ context.Context, problemID uint) error {
	p.deletedProblemIDs = append(p.deletedProblemIDs, problemID)
	return nil
}

func (p *recordingSyncProducer) EnqueueUserScoreSync(_ context.Context, _ uint) error {
	return nil
}

func TestNewProblemServiceInjectsSyncProducer(t *testing.T) {
	producer := &recordingSyncProducer{}
	service := NewProblemService(nil, nil, producer)

	service.enqueueProblemUpsert(context.Background(), 41, "test")
	service.enqueueProblemDelete(context.Background(), 42, "test")

	if len(producer.upsertedProblemIDs) != 1 || producer.upsertedProblemIDs[0] != 41 {
		t.Fatalf("upsert calls = %v, want [41]", producer.upsertedProblemIDs)
	}
	if len(producer.deletedProblemIDs) != 1 || producer.deletedProblemIDs[0] != 42 {
		t.Fatalf("delete calls = %v, want [42]", producer.deletedProblemIDs)
	}
}
