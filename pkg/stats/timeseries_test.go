package stats

import (
	"testing"
	"time"
)

func TestTimeSeriesRollingWindow(t *testing.T) {
	ts := NewTimeSeries(2)
	ts.Add(DataPoint{Timestamp: time.Unix(1, 0), PacketsPerSec: 1})
	ts.Add(DataPoint{Timestamp: time.Unix(2, 0), PacketsPerSec: 2})
	ts.Add(DataPoint{Timestamp: time.Unix(3, 0), PacketsPerSec: 3})

	got := ts.GetAll()
	if len(got) != 2 {
		t.Fatalf("len(GetAll()) = %d, want 2", len(got))
	}
	if got[0].PacketsPerSec != 2 || got[1].PacketsPerSec != 3 {
		t.Fatalf("unexpected rolling window: %+v", got)
	}
}

func TestTimeSeriesGetLastHandlesNonPositiveN(t *testing.T) {
	ts := NewTimeSeries(2)
	ts.Add(DataPoint{PacketsPerSec: 1})

	for _, n := range []int{0, -1} {
		got := ts.GetLast(n)
		if len(got) != 0 {
			t.Fatalf("GetLast(%d) len = %d, want 0", n, len(got))
		}
	}
}

func TestConnectionTableEvictsOldest(t *testing.T) {
	ct := NewConnectionTable(2)
	first := ConnectionKey{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Proto: "TCP"}
	second := ConnectionKey{SrcIP: "10.0.0.3", DstIP: "10.0.0.4", SrcPort: 3, DstPort: 4, Proto: "TCP"}
	third := ConnectionKey{SrcIP: "10.0.0.5", DstIP: "10.0.0.6", SrcPort: 5, DstPort: 6, Proto: "TCP"}

	ct.Update(first, 100)
	time.Sleep(time.Millisecond)
	ct.Update(second, 100)
	time.Sleep(time.Millisecond)
	ct.Update(third, 100)

	entries := ct.GetAll()
	if len(entries) != 2 {
		t.Fatalf("len(GetAll()) = %d, want 2", len(entries))
	}
	for _, entry := range entries {
		if entry.SrcIP == first.SrcIP {
			t.Fatalf("oldest connection was not evicted: %+v", entry)
		}
	}
}
