package main

import (
	"testing"
	"time"
)

// TestRepairBackoff pins the schedule the guard relies on. The progression is
// what stops the recorded repair storm: 2m after one failed repair, doubling
// from there, capped so a long outage still retries regularly.
func TestRepairBackoff(t *testing.T) {
	cases := []struct {
		fails int
		want  time.Duration
	}{
		{0, 0},
		{-1, 0},
		{1, 2 * time.Minute},
		{2, 4 * time.Minute},
		{3, 8 * time.Minute},
		{4, 16 * time.Minute},
		{5, repairBackoffMax},
		{10, repairBackoffMax},
		{100, repairBackoffMax},
	}
	for _, c := range cases {
		if got := repairBackoff(c.fails); got != c.want {
			t.Errorf("repairBackoff(%d) = %v, want %v", c.fails, got, c.want)
		}
	}
}

// TestRepairBackoffMonotonic guards the property that matters in practice: a
// longer streak of failures never shortens the wait.
func TestRepairBackoffMonotonic(t *testing.T) {
	prev := time.Duration(0)
	for fails := 0; fails <= 12; fails++ {
		got := repairBackoff(fails)
		if got < prev {
			t.Fatalf("repairBackoff(%d) = %v shrank below the previous %v", fails, got, prev)
		}
		prev = got
	}
}

// TestInsuranceWaitAdvancesSchedule pins that the counter advances every call
// and that the scheduled time moves forward with it. A version that forgot to
// increment would silently disable the backoff; one that never moved the
// deadline would stall the session forever.
func TestInsuranceWaitAdvancesSchedule(t *testing.T) {
	var (
		fails    int
		nextAt   time.Time
		lastWait time.Duration
	)
	for i := 1; i <= 5; i++ {
		wait := insuranceWait(&fails, &nextAt)
		if fails != i {
			t.Fatalf("after call %d the counter is %d, want %d", i, fails, i)
		}
		if want := repairBackoff(i); wait != want {
			t.Fatalf("call %d scheduled %v, want %v", i, wait, want)
		}
		if wait < lastWait {
			t.Fatalf("call %d scheduled %v, shorter than the previous %v", i, wait, lastWait)
		}
		if nextAt.Before(time.Now()) {
			t.Fatalf("call %d scheduled a deadline in the past", i)
		}
		lastWait = wait
	}
	// Capped, so the fifth call is already at the ceiling.
	if lastWait != repairBackoffMax {
		t.Fatalf("final wait = %v, want the %v ceiling", lastWait, repairBackoffMax)
	}
}

// TestInsuranceWaitZeroMeansNoHold pins that a fresh schedule (counter starts
// at zero before the first call) always produces a positive wait, so the guard
// cannot accidentally skip its own backoff on the first failure.
func TestInsuranceWaitZeroMeansNoHold(t *testing.T) {
	if got := repairBackoff(0); got != 0 {
		t.Fatalf("repairBackoff(0) = %v, want 0 (no pending repair)", got)
	}
	var (
		fails int
		next  time.Time
	)
	if wait := insuranceWait(&fails, &next); wait <= 0 {
		t.Fatalf("first insuranceWait returned %v, want a positive wait", wait)
	}
}
