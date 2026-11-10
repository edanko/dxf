package genetic

import (
	"math/rand"
	"testing"
)

func TestFloatDNAValidationAndMutation(t *testing.T) {
	dna, err := NewFloatDNA([]float64{0.25, 0.75})
	if err != nil {
		t.Fatal(err)
	}
	if dna.Len() != 2 || dna.Gene(0) != 0.25 {
		t.Fatalf("unexpected DNA: %s", dna)
	}
	if err := dna.FlipMutateAt(0); err != nil {
		t.Fatal(err)
	}
	if dna.Gene(0) != 0.75 {
		t.Fatalf("flip mutation = %v, want 0.75", dna.Gene(0))
	}
	if _, err := NewFloatDNA([]float64{1.5}); err == nil {
		t.Fatal("expected out-of-range error")
	}
}

func TestBitDNAIntegerMappings(t *testing.T) {
	bits := NewBitDNA([]bool{true, false, true})
	if bits.Gene(0) != 1 || bits.Gene(1) != 0 {
		t.Fatalf("unexpected bit genes: %v", dnaGenes(bits))
	}
	if err := bits.FlipMutateAt(1); err != nil {
		t.Fatal(err)
	}
	if !bits.IsValid() || bits.Gene(1) != 1 {
		t.Fatal("bit flip mutation failed")
	}

	perm := NewUniqueIntDNARange(4)
	if !perm.IsValid() {
		t.Fatal("range DNA should be valid")
	}
	if err := perm.FlipMutateAt(0); err == nil {
		t.Fatal("unique DNA flip should fail")
	}
	if _, err := NewUniqueIntDNA([]int{0, 0, 1}); err == nil {
		t.Fatal("expected duplicate error")
	}

	ints, err := NewIntegerDNA([]int{0, 2, 1}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if err := ints.FlipMutateAt(0); err != nil {
		t.Fatal(err)
	}
	if ints.Gene(0) != 2 {
		t.Fatalf("integer flip = %v, want 2", ints.Gene(0))
	}
}

func TestMateOrderedKeepsPermutationValid(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	a := RandomUniqueIntDNA(rng, 10)
	b := RandomUniqueIntDNA(rng, 10)
	MateOrderedCX{}.Recombine(rng, a, b)
	if !a.IsValid() || !b.IsValid() {
		t.Fatalf("ordered crossover produced invalid permutations: %v %v", a, b)
	}
}

func TestOptimizerMaximizesBitFitness(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	opt := NewOptimizer(EvaluatorFunc(func(dna DNA) float64 {
		total := 0.0
		for i := 0; i < dna.Len(); i++ {
			total += dna.Gene(i)
		}
		return total
	}), 120)
	opt.Rand = rng
	opt.MaxFitness = 8
	opt.CrossoverRate = 0.8
	opt.MutationRate = 0.05
	opt.Elitism = 2
	if err := opt.AddCandidates(RandomBitDNAMany(rng, 20, 8)); err != nil {
		t.Fatal(err)
	}
	if err := opt.Execute(nil); err != nil {
		t.Fatal(err)
	}
	if opt.BestFitness != 8 {
		t.Fatalf("best fitness = %v, want 8 (%s)", opt.BestFitness, opt)
	}
	if opt.BestDNA == nil || opt.BestDNA.Len() != 8 {
		t.Fatalf("unexpected best DNA: %v", opt.BestDNA)
	}
	if len(opt.Log.Entries) == 0 {
		t.Fatal("expected generation log entries")
	}
}

func TestThresholdFilterAndSelections(t *testing.T) {
	candidates := []DNA{
		&BitDNA{dnaBase: dnaBase{fitness: 1, hasFitness: true}, data: []bool{true}},
		&BitDNA{dnaBase: dnaBase{fitness: 10, hasFitness: true}, data: []bool{true}},
		&BitDNA{dnaBase: dnaBase{fitness: 20, hasFitness: true}, data: []bool{true}},
	}
	filtered := ThresholdFilter(candidates, 20, 0.4)
	if len(filtered) != 2 {
		t.Fatalf("threshold filter kept %d candidates, want 2", len(filtered))
	}

	rng := rand.New(rand.NewSource(3))
	roulette := &RouletteSelection{}
	roulette.Reset(candidates)
	if got := len(roulette.Pick(rng, 5)); got != 5 {
		t.Fatalf("roulette picked %d, want 5", got)
	}

	ranked := &RankBasedSelection{}
	ranked.Reset(candidates)
	if got := len(ranked.Pick(rng, 5)); got != 5 {
		t.Fatalf("rank selection picked %d, want 5", got)
	}

	tournament := &TournamentSelection{Candidates: 2}
	tournament.Reset(candidates)
	picked := tournament.Pick(rng, 3)
	if len(picked) != 3 {
		t.Fatalf("tournament picked %d, want 3", len(picked))
	}
	for _, dna := range picked {
		if dna.Fitness() != 20 {
			t.Fatalf("tournament picked fitness %v, want 20", dna.Fitness())
		}
	}
}

func TestOptimizerRequiresCandidates(t *testing.T) {
	opt := NewOptimizer(EvaluatorFunc(func(DNA) float64 { return 0 }), 1)
	if err := opt.Execute(nil); err == nil {
		t.Fatal("expected error for empty population")
	}
}
