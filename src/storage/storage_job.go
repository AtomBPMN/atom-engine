/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"atom-engine/src/core/models"

	"github.com/dgraph-io/badger/v3"
)

// Job storage methods

// SaveJob saves job to storage
func (bs *BadgerStorage) SaveJob(ctx context.Context, job *models.Job) error {
	key := fmt.Sprintf("job:%s", job.ID)
	return bs.saveJSON(key, job)
}

// GetJob gets job from storage
func (bs *BadgerStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	key := fmt.Sprintf("job:%s", jobID)
	var job models.Job

	err := bs.loadJSON(key, &job)
	if err != nil {
		if err.Error() == fmt.Sprintf("key not found: %s", key) {
			return nil, nil // Job not found
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// ListJobsByType lists jobs by type and status
func (bs *BadgerStorage) ListJobsByType(ctx context.Context, jobType string, status models.JobStatus, limit int) ([]*models.Job, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var jobs []*models.Job
	prefix := []byte("job:")

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix) && (limit <= 0 || count < limit); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var job models.Job
				if err := json.Unmarshal(val, &job); err != nil {
					return err
				}

				// Filter by type if specified
				if jobType != "" && job.Type != jobType {
					return nil
				}

				// Filter by status if specified
				if status != "" && job.Status != status {
					return nil
				}

				jobs = append(jobs, &job)
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	return jobs, nil
}
