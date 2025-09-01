package main

import "testing"

func TestAppendMatchingSign(t *testing.T) {
	s := NewStore()
	_, _ = s.Apply(5)
	_, state := s.Apply(7)
	want := []int{5, 7}
	if !equal(state, want) {
		t.Fatalf("got %v, want %v", state, want)
	}
}

func TestOppositeSignPartialConsume(t *testing.T) {
	s := NewStore()
	_, _ = s.Apply(5)
	_, _ = s.Apply(10)
	_, state := s.Apply(-6) // expect [9]
	want := []int{9}
	if !equal(state, want) {
		t.Fatalf("got %v, want %v", state, want)
	}
}

func TestOppositeSignExactConsumeElement(t *testing.T) {
	s := NewStore()
	_, _ = s.Apply(5)
	_, _ = s.Apply(10)
	_, state := s.Apply(-5) // consume first 5 exactly -> [10]
	want := []int{10}
	if !equal(state, want) {
		t.Fatalf("got %v, want %v", state, want)
	}
}

func TestOppositeSignExceedTotalFlipSign(t *testing.T) {
	s := NewStore()
	_, _ = s.Apply(4)
	_, _ = s.Apply(3)  // total 7
	_, state := s.Apply(-10) // exhaust list, remainder -3 -> [-3]
	want := []int{-3}
	if !equal(state, want) {
		t.Fatalf("got %v, want %v", state, want)
	}
}

func TestZeroNoOp(t *testing.T) {
	s := NewStore()
	_, _ = s.Apply(5)
	_, state := s.Apply(0)
	want := []int{5}
	if !equal(state, want) {
		t.Fatalf("zero should be no-op; got %v, want %v", state, want)
	}
}

func TestStartWithNegative(t *testing.T) {
	s := NewStore()
	_, _ = s.Apply(-8)
	_, state := s.Apply(-2)
	want := []int{-8, -2}
	if !equal(state, want) {
		t.Fatalf("got %v, want %v", state, want)
	}
	_, state = s.Apply(5) // opposite; consume 5 from -8 -> [-3, -2]
	want = []int{-3, -2}
	if !equal(state, want) {
		t.Fatalf("got %v, want %v", state, want)
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
