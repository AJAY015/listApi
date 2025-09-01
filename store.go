package main

import (
	"fmt"
	"log"
	"math"
	"sync"
)

type Store struct {
	mu    sync.Mutex
	items []int // FIFO queue of non-zero integers
}

func NewStore() *Store {
	return &Store{items: make([]int, 0)}
}

func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = s.items[:0]
	log.Printf("[reset] list -> %v", s.items)
}

func (s *Store) Snapshot() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]int, len(s.items))
	copy(out, s.items)
	return out
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}

// Apply applies the number n to the store following the rules:
// - n == 0 => no-op
// - if list is empty => append n
// - if sign(n) == sign(list) => append n (FIFO append)
// - else => consume from the head (FIFO) by |n| until exhausted; if remainder remains after
//           fully consuming the list, append the remainder with sign(n) (sign may flip).
//
// Returns a human-readable action and the updated list snapshot.
func (s *Store) Apply(n int) (string, []int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n == 0 {
		log.Printf("[apply] input=0 -> no-op, list=%v", s.items)
		return "ignored zero (no-op)", snapshot(s.items)
	}

	nsgn := sign(n)
	if len(s.items) == 0 {
		s.items = append(s.items, n)
		log.Printf("[apply] list empty, appended %d -> %v", n, s.items)
		return fmt.Sprintf("list empty, appended %d", n), snapshot(s.items)
	}

	// Determine current list sign from the first non-zero element.
	// Given our invariant, all items should share sign, but we still guard.
	listSign := sign(s.items[0])
	if listSign == 0 {
		// If somehow first is zero (shouldn't happen), compact zeros and re-evaluate.
		comp := compactNonZero(s.items)
		s.items = comp
		if len(s.items) == 0 {
			s.items = append(s.items, n)
			log.Printf("[apply] after compact empty, appended %d -> %v", n, s.items)
			return fmt.Sprintf("list empty after compaction, appended %d", n), snapshot(s.items)
		}
		listSign = sign(s.items[0])
	}

	if nsgn == listSign {
		s.items = append(s.items, n)
		log.Printf("[apply] same sign (%+d), appended %d -> %v", listSign, n, s.items)
		return fmt.Sprintf("same sign, appended %d", n), snapshot(s.items)
	}

	// Opposite sign: consume from head by |n| (FIFO).
	rem := int(math.Abs(float64(n)))
	newItems := make([]int, 0, len(s.items))

	for i := 0; i < len(s.items) && rem > 0; i++ {
		x := s.items[i]
		ax := int(math.Abs(float64(x)))

		switch {
		case rem > ax:
			// consume whole element
			rem -= ax
			log.Printf("[consume] %d consumed fully by %d, remainder=%d", x, n, rem)
		case rem == ax:
			// consume exactly, drop element
			rem = 0
			log.Printf("[consume] %d consumed exactly by %d", x, n)
			// elements after i+1 (if any) will be appended after loop
			// but since rem==0, we'll append them below.
			// (we do nothing here: we drop x)
			// continue
		default: // rem < ax
			// partially consume; keep leftover with original list sign
			newVal := (ax - rem) * listSign
			rem = 0
			newItems = append(newItems, newVal)
			log.Printf("[consume] %d partially consumed by %d -> leftover %d", x, n, newVal)
			// append rest unchanged
			for j := i + 1; j < len(s.items); j++ {
				newItems = append(newItems, s.items[j])
			}
		}
	}

	// If we exited loop with rem > 0, we consumed whole list. Append remainder with input sign.
	if rem > 0 {
		leftover := rem * nsgn
		newItems = append(newItems, leftover)
		log.Printf("[flip] list exhausted; appended remainder %d (sign flip possible)", leftover)
	} else if len(newItems) == 0 {
		// rem == 0 and we consumed up to some point; append any tail not yet appended.
		// (This happens if we consumed elements exactly and broke due to rem==0.)
		for i := 0; i < len(s.items); i++ {
			x := s.items[i]
			if x != 0 {
				newItems = append(newItems, x)
			}
		}
	}

	// Compact zeroes just in case (shouldn’t be necessary with logic above).
	newItems = compactNonZero(newItems)
	s.items = newItems
	log.Printf("[apply] opposite sign, input=%d -> %v", n, s.items)
	return fmt.Sprintf("opposite sign, applied %d", n), snapshot(s.items)
}

func compactNonZero(a []int) []int {
	out := a[:0]
	for _, v := range a {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}

func snapshot(a []int) []int {
	out := make([]int, len(a))
	copy(out, a)
	return out
}
