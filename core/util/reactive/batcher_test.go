//go:build goexperiment.synctest

/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reactive

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"testing/synctest"
	"time"
)

var (
	opDelay     = time.Nanosecond
	batchWait   = 3 * time.Second
	processTime = 5 * time.Second
)

func TestThrottleBatcher_LimitReached(t *testing.T) {
	synctest.Run(func() {
		var b = NewThrottleBatcher[int](3, batchWait)
		var count int

		// Put 6 items (> softLimit) to batcher before start
		// All of that will be process right after b.Start()
		for i := range 6 {
			b.Put(i)
			time.Sleep(opDelay)
		}

		b.Start(func(items []int) {
			count++
			switch count {
			case 1:
				assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, items)
			case 2:
				assert.Equal(t, []int{10, 11, 12}, items)
			case 3:
				assert.Equal(t, []int{13, 14}, items)
			default:
				t.Error("unexpected this call")
			}
		})
		time.Sleep(opDelay)

		// Put 5 items (softLimit < 5 < 2*softLimit) after batcher Start
		// Them will be divided into 2 batch:
		// * one size of 3
		// * one size of 2 - flush when b.Stop()
		for i := range 5 {
			b.Put(10 + i)
			time.Sleep(opDelay)
		}

		b.Stop()

		assert.Equal(t, 3, count)
	})
}

func TestThrottleBatcher_TimeElapsed(t *testing.T) {
	synctest.Run(func() {
		var b = NewThrottleBatcher[int](5, batchWait)
		var count int

		// Put 2 items (< softLimit) to batcher before start
		// After Start, it will wait for more items
		for i := range 2 {
			b.Put(i)
			time.Sleep(opDelay)
		}

		b.Start(func(items []int) {
			count++
			switch count {
			case 1:
				assert.Equal(t, []int{0, 1, 2, 3}, items)
			case 2:
				assert.Equal(t, []int{10, 11, 12}, items)
			case 3:
				assert.Equal(t, []int{20, 21}, items)
			default:
				t.Error("unexpected this call")
			}
		})
		time.Sleep(opDelay)

		// Add more 2 items, still not reach the batch size
		for i := range 2 {
			b.Put(2 + i)
			time.Sleep(opDelay)
		}

		// Waits for batchWait duration to elapse, all queued items are processed
		time.Sleep(batchWait)

		for i := range 3 {
			b.Put(10 + i)
			time.Sleep(opDelay)
		}
		// Waits for batchWait duration to elapse, all queued items are processed
		time.Sleep(batchWait)

		// Last 2 items will be flush when b.Stop()
		for i := range 2 {
			b.Put(20 + i)
			time.Sleep(opDelay)
		}
		b.Stop()

		assert.Equal(t, 3, count)
	})
}

func TestThrottleBatcher_SlowProcess(t *testing.T) {
	synctest.Run(func() {
		var b = NewThrottleBatcher[int](3, batchWait)
		var count int

		b.Start(func(items []int) {
			count++
			switch count {
			case 1:
				assert.Equal(t, []int{0, 1, 2}, items)
			case 2:
				assert.Equal(t, []int{3, 4, 10, 11, 12}, items)
			case 3:
				assert.Equal(t, []int{20, 21}, items)
			default:
				t.Error("unexpected this call")
			}
			time.Sleep(processTime)
		})

		// Put 5 items - first 3 of them will be process - take 5s
		// 2 items left
		for i := range 5 {
			b.Put(i)
			time.Sleep(opDelay)
		}

		// batchWait duration pass, but 2 remaining items will NOT be processed
		// because b is busy with first batch
		time.Sleep(batchWait)

		// Put more 3 items -> total 5 in queue
		for i := range 3 {
			b.Put(10 + i)
			time.Sleep(opDelay)
		}
		// wait more 3s, now b is finish first batch.
		// second batch of all 5 items in queue is processing
		time.Sleep(batchWait)

		for i := range 2 {
			b.Put(20 + i)
			time.Sleep(opDelay)
		}
		b.Stop()

		assert.Equal(t, 3, count)
	})
}

func TestThrottleBatcher_StopEmpty(t *testing.T) {
	synctest.Run(func() {
		var b = NewThrottleBatcher[int](5, batchWait)

		b.Start(func(items []int) {
			t.Error("unexpected this call")
		})
		b.Stop()

		synctest.Wait()
	})
}
