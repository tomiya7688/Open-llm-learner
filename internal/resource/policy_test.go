package resource

import "testing"

func TestAutoBudgetPreservesHeadroom(t *testing.T) {
	p := Policy{}
	p.Normalize()
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	b, err := ComputeBudget(1000, p.RAM)
	if err != nil {
		t.Fatal(err)
	}
	if b.UsableBytes != 900 || b.ReservedBytes != 100 {
		t.Fatalf("unexpected budget: %+v", b)
	}
}

func TestExplicitReserveWins(t *testing.T) {
	fraction := 0.95
	reserve := uint64(300)
	b, err := ComputeBudget(1000, MemoryLimit{Mode: "auto", MaxFraction: &fraction, MinFreeBytes: &reserve})
	if err != nil {
		t.Fatal(err)
	}
	if b.UsableBytes != 700 || b.ReservedBytes != 300 {
		t.Fatalf("unexpected budget: %+v", b)
	}
}

func TestAutoRejectsFullUtilization(t *testing.T) {
	fraction := 1.0
	_, err := ComputeBudget(1000, MemoryLimit{Mode: "auto", MaxFraction: &fraction})
	if err == nil {
		t.Fatal("expected auto mode max_fraction=1 to be rejected")
	}
}
