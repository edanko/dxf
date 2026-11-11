// Package genetic implements a small, reusable genetic algorithm engine.
//
// The API mirrors the structure of ezdxf's genetic_algorithm add-on: DNA
// containers, mutation and recombination strategies, selection strategies and
// an optimizer which always maximizes fitness. Callers supply an Evaluator and
// an initial population.
package genetic

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/edanko/dxf/addons"
)

// DNA is a mutable candidate solution.
type DNA interface {
	// Len returns the number of genes.
	Len() int
	// Gene returns the gene at index as a float64. Integer and bit DNA types
	// map their native representation onto this scalar view.
	Gene(index int) float64
	// SetGene changes a gene and clears any cached fitness.
	SetGene(index int, value float64)
	// Copy returns a deep copy of the DNA, including cached fitness.
	Copy() DNA
	// Fitness returns the cached fitness value. Use HasFitness to check
	// whether a value has been evaluated.
	Fitness() float64
	// HasFitness reports whether the cached fitness is valid.
	HasFitness() bool
	// SetFitness caches a fitness value.
	SetFitness(fitness float64)
	// IsValid reports whether the genes satisfy the DNA type constraints.
	IsValid() bool
	// ResetGenes replaces all genes and clears cached fitness.
	ResetGenes(values []float64) error
	// FlipMutateAt applies the type-specific flip mutation.
	FlipMutateAt(index int) error
	// String returns a human readable representation.
	String() string
}

type dnaBase struct {
	fitness    float64
	hasFitness bool
}

func (d *dnaBase) Fitness() float64     { return d.fitness }
func (d *dnaBase) HasFitness() bool     { return d.hasFitness }
func (d *dnaBase) SetFitness(v float64) { d.fitness = v; d.hasFitness = true }
func (d *dnaBase) taint()               { d.hasFitness = false }
func (d *dnaBase) copyBase() dnaBase    { return dnaBase{fitness: d.fitness, hasFitness: d.hasFitness} }
func formatFitness(f float64, ok bool) string {
	if !ok {
		return "fitness=none"
	}
	return fmt.Sprintf("fitness=%.4f", f)
}

// FloatDNA stores arbitrary float values constrained to [0, 1].
type FloatDNA struct {
	dnaBase
	data []float64
}

// NewFloatDNA creates float DNA and validates that every value is in [0, 1].
func NewFloatDNA(values []float64) (*FloatDNA, error) {
	data := append([]float64(nil), values...)
	dna := &FloatDNA{data: data}
	if !dna.IsValid() {
		return nil, errors.New("genetic: float DNA values must be in [0, 1]")
	}
	return dna, nil
}

// RandomFloatDNA creates random float DNA using rng.
func RandomFloatDNA(rng *rand.Rand, length int) *FloatDNA {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}
	data := make([]float64, length)
	for i := range data {
		data[i] = rng.Float64()
	}
	return &FloatDNA{data: data}
}

// RandomFloatDNAMany creates count random float DNA strands.
func RandomFloatDNAMany(rng *rand.Rand, count, length int) []DNA {
	rng = normalizeRNG(rng)
	out := make([]DNA, count)
	for i := range out {
		out[i] = RandomFloatDNA(rng, length)
	}
	return out
}

func (d *FloatDNA) Len() int               { return len(d.data) }
func (d *FloatDNA) Gene(index int) float64 { return d.data[index] }
func (d *FloatDNA) SetGene(index int, value float64) {
	d.data[index] = value
	d.taint()
}
func (d *FloatDNA) Copy() DNA {
	return &FloatDNA{dnaBase: d.copyBase(), data: append([]float64(nil), d.data...)}
}
func (d *FloatDNA) IsValid() bool {
	for _, v := range d.data {
		if v < 0 || v > 1 {
			return false
		}
	}
	return true
}
func (d *FloatDNA) ResetGenes(values []float64) error {
	d.data = append(d.data[:0], values...)
	if !d.IsValid() {
		return errors.New("genetic: float DNA values must be in [0, 1]")
	}
	d.taint()
	return nil
}
func (d *FloatDNA) FlipMutateAt(index int) error {
	d.data[index] = 1 - d.data[index]
	d.taint()
	return nil
}
func (d *FloatDNA) String() string {
	return fmt.Sprintf("FloatDNA(%v, %s)", d.data, formatFitness(d.fitness, d.hasFitness))
}

// BitDNA stores boolean genes.
type BitDNA struct {
	dnaBase
	data []bool
}

// NewBitDNA creates bit DNA; any non-zero input becomes true.
func NewBitDNA(values []bool) *BitDNA {
	return &BitDNA{data: append([]bool(nil), values...)}
}

// RandomBitDNA creates random bit DNA using rng.
func RandomBitDNA(rng *rand.Rand, length int) *BitDNA {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}
	data := make([]bool, length)
	for i := range data {
		data[i] = rng.Intn(2) == 1
	}
	return &BitDNA{data: data}
}

// RandomBitDNAMany creates count random bit DNA strands.
func RandomBitDNAMany(rng *rand.Rand, count, length int) []DNA {
	rng = normalizeRNG(rng)
	out := make([]DNA, count)
	for i := range out {
		out[i] = RandomBitDNA(rng, length)
	}
	return out
}

func (d *BitDNA) Len() int { return len(d.data) }
func (d *BitDNA) Gene(index int) float64 {
	if d.data[index] {
		return 1
	}
	return 0
}
func (d *BitDNA) SetGene(index int, value float64) {
	d.data[index] = value != 0
	d.taint()
}
func (d *BitDNA) Copy() DNA {
	return &BitDNA{dnaBase: d.copyBase(), data: append([]bool(nil), d.data...)}
}
func (d *BitDNA) IsValid() bool { return true }
func (d *BitDNA) ResetGenes(values []float64) error {
	d.data = d.data[:0]
	for _, v := range values {
		d.data = append(d.data, v != 0)
	}
	d.taint()
	return nil
}
func (d *BitDNA) FlipMutateAt(index int) error {
	d.data[index] = !d.data[index]
	d.taint()
	return nil
}
func (d *BitDNA) String() string {
	values := make([]int, len(d.data))
	for i, v := range d.data {
		if v {
			values[i] = 1
		}
	}
	return fmt.Sprintf("BitDNA(%v, %s)", values, formatFitness(d.fitness, d.hasFitness))
}

// UniqueIntDNA stores a permutation of the integers 0..len-1.
type UniqueIntDNA struct {
	dnaBase
	data []int
}

// NewUniqueIntDNA creates a permutation. An integer argument creates
// 0..n-1; a slice argument must contain unique values.
func NewUniqueIntDNA(values []int) (*UniqueIntDNA, error) {
	data := append([]int(nil), values...)
	dna := &UniqueIntDNA{data: data}
	if !dna.IsValid() {
		return nil, errors.New("genetic: unique int DNA values must be distinct")
	}
	return dna, nil
}

// NewUniqueIntDNARange creates [0, 1, ..., length-1].
func NewUniqueIntDNARange(length int) *UniqueIntDNA {
	data := make([]int, length)
	for i := range data {
		data[i] = i
	}
	return &UniqueIntDNA{data: data}
}

// RandomUniqueIntDNA creates a shuffled permutation using rng.
func RandomUniqueIntDNA(rng *rand.Rand, length int) *UniqueIntDNA {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}
	dna := NewUniqueIntDNARange(length)
	rng.Shuffle(length, func(i, j int) { dna.data[i], dna.data[j] = dna.data[j], dna.data[i] })
	return dna
}

// RandomUniqueIntDNAMany creates count shuffled permutations.
func RandomUniqueIntDNAMany(rng *rand.Rand, count, length int) []DNA {
	rng = normalizeRNG(rng)
	out := make([]DNA, count)
	for i := range out {
		out[i] = RandomUniqueIntDNA(rng, length)
	}
	return out
}

func (d *UniqueIntDNA) Len() int               { return len(d.data) }
func (d *UniqueIntDNA) Gene(index int) float64 { return float64(d.data[index]) }
func (d *UniqueIntDNA) SetGene(index int, value float64) {
	d.data[index] = int(math.Round(value))
	d.taint()
}
func (d *UniqueIntDNA) Copy() DNA {
	return &UniqueIntDNA{dnaBase: d.copyBase(), data: append([]int(nil), d.data...)}
}
func (d *UniqueIntDNA) IsValid() bool {
	seen := make(map[int]struct{}, len(d.data))
	for _, v := range d.data {
		if _, ok := seen[v]; ok {
			return false
		}
		seen[v] = struct{}{}
	}
	return true
}
func (d *UniqueIntDNA) ResetGenes(values []float64) error {
	data := make([]int, len(values))
	for i, v := range values {
		data[i] = int(math.Round(v))
	}
	d.data = data
	if !d.IsValid() {
		return errors.New("genetic: unique int DNA values must be distinct")
	}
	d.taint()
	return nil
}
func (d *UniqueIntDNA) FlipMutateAt(index int) error {
	return errors.New("genetic: flip mutation not supported for unique int DNA")
}
func (d *UniqueIntDNA) String() string {
	return fmt.Sprintf("UniqueIntDNA(%v, %s)", d.data, formatFitness(d.fitness, d.hasFitness))
}

// IntegerDNA stores integer values in the range [0, max-1].
type IntegerDNA struct {
	dnaBase
	data []int
	max  int
}

// NewIntegerDNA creates integer DNA with values in [0, max).
func NewIntegerDNA(values []int, max int) (*IntegerDNA, error) {
	dna := &IntegerDNA{data: append([]int(nil), values...), max: max}
	if !dna.IsValid() {
		return nil, fmt.Errorf("genetic: integer DNA values must be in [0, %d)", max)
	}
	return dna, nil
}

// RandomIntegerDNA creates random integer DNA using rng.
func RandomIntegerDNA(rng *rand.Rand, length, max int) *IntegerDNA {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}
	data := make([]int, length)
	for i := range data {
		data[i] = rng.Intn(max)
	}
	return &IntegerDNA{data: data, max: max}
}

// RandomIntegerDNAMany creates count random integer DNA strands.
func RandomIntegerDNAMany(rng *rand.Rand, count, length, max int) []DNA {
	rng = normalizeRNG(rng)
	out := make([]DNA, count)
	for i := range out {
		out[i] = RandomIntegerDNA(rng, length, max)
	}
	return out
}

func (d *IntegerDNA) Len() int               { return len(d.data) }
func (d *IntegerDNA) Gene(index int) float64 { return float64(d.data[index]) }
func (d *IntegerDNA) SetGene(index int, value float64) {
	d.data[index] = int(math.Round(value))
	d.taint()
}
func (d *IntegerDNA) Copy() DNA {
	return &IntegerDNA{dnaBase: d.copyBase(), data: append([]int(nil), d.data...), max: d.max}
}
func (d *IntegerDNA) IsValid() bool {
	for _, v := range d.data {
		if v < 0 || v >= d.max {
			return false
		}
	}
	return true
}
func (d *IntegerDNA) ResetGenes(values []float64) error {
	data := make([]int, len(values))
	for i, v := range values {
		data[i] = int(math.Round(v))
	}
	d.data = data
	if !d.IsValid() {
		return fmt.Errorf("genetic: integer DNA values must be in [0, %d)", d.max)
	}
	d.taint()
	return nil
}
func (d *IntegerDNA) FlipMutateAt(index int) error {
	d.data[index] = d.max - d.data[index] - 1
	d.taint()
	return nil
}
func (d *IntegerDNA) String() string {
	return fmt.Sprintf("IntegerDNA(%v, %d, %s)", d.data, d.max, formatFitness(d.fitness, d.hasFitness))
}

// Mutate is a mutation strategy.
type Mutate interface {
	Mutate(rng *rand.Rand, dna DNA, rate float64)
}

func normalizeRNG(rng *rand.Rand) *rand.Rand {
	if rng != nil {
		return rng
	}
	return rand.New(rand.NewSource(0))
}

// FlipMutate flips each gene independently with probability rate.
type FlipMutate struct{}

func (FlipMutate) Mutate(rng *rand.Rand, dna DNA, rate float64) {
	rng = normalizeRNG(rng)
	for i := 0; i < dna.Len(); i++ {
		if rng.Float64() < rate {
			_ = dna.FlipMutateAt(i)
		}
	}
}

// NeighborSwapMutate swaps neighboring genes with probability rate.
type NeighborSwapMutate struct{}

func (NeighborSwapMutate) Mutate(rng *rand.Rand, dna DNA, rate float64) {
	rng = normalizeRNG(rng)
	for i := 1; i < dna.Len(); i++ {
		if rng.Float64() < rate {
			swap(dna, i-1, i)
		}
	}
}

// RandomSwapMutate swaps each gene with a random gene with probability rate.
type RandomSwapMutate struct{}

func (RandomSwapMutate) Mutate(rng *rand.Rand, dna DNA, rate float64) {
	rng = normalizeRNG(rng)
	length := dna.Len()
	if length < 2 {
		return
	}
	for i := range length {
		if rng.Float64() < rate {
			j := rng.Intn(length)
			if j == i {
				j = (j + 1) % length
			}
			swap(dna, i, j)
		}
	}
}

// ReverseMutate reverses a window of genes.
type ReverseMutate struct{ Bits int }

// NewReverseMutate creates a reverse mutation with a window of at least one.
func NewReverseMutate(bits int) *ReverseMutate {
	if bits < 1 {
		bits = 1
	}
	return &ReverseMutate{Bits: bits}
}

func (m *ReverseMutate) Mutate(rng *rand.Rand, dna DNA, rate float64) {
	rng = normalizeRNG(rng)
	length := dna.Len()
	if length < m.Bits || m.Bits < 1 || length == 0 {
		return
	}
	if rng.Float64() >= rate*float64(length)/float64(m.Bits) {
		return
	}
	start := rng.Intn(length - m.Bits + 1)
	for i, j := start, start+m.Bits-1; i < j; i, j = i+1, j-1 {
		swap(dna, i, j)
	}
}

// ScrambleMutate shuffles a window of genes.
type ScrambleMutate struct{ Bits int }

// NewScrambleMutate creates a scramble mutation with a window of at least one.
func NewScrambleMutate(bits int) *ScrambleMutate {
	if bits < 1 {
		bits = 1
	}
	return &ScrambleMutate{Bits: bits}
}

func (m *ScrambleMutate) Mutate(rng *rand.Rand, dna DNA, rate float64) {
	rng = normalizeRNG(rng)
	length := dna.Len()
	if length < m.Bits || m.Bits < 1 || length == 0 {
		return
	}
	if rng.Float64() >= rate*float64(length)/float64(m.Bits) {
		return
	}
	start := rng.Intn(length - m.Bits + 1)
	for i := start + m.Bits - 1; i > start; i-- {
		j := start + rng.Intn(i-start+1)
		swap(dna, i, j)
	}
}

func swap(dna DNA, i, j int) {
	a, b := dna.Gene(i), dna.Gene(j)
	dna.SetGene(i, b)
	dna.SetGene(j, a)
}

// Mate is a recombination strategy.
type Mate interface {
	Recombine(rng *rand.Rand, dna1, dna2 DNA)
}

// Mate1pCX performs one-point crossover.
type Mate1pCX struct{}

func (Mate1pCX) Recombine(rng *rand.Rand, dna1, dna2 DNA) {
	rng = normalizeRNG(rng)
	length := dna1.Len()
	if length < 2 {
		return
	}
	index := rng.Intn(length)
	recombineRange(dna1, dna2, index, length)
}

// Mate2pCX performs two-point crossover.
type Mate2pCX struct{}

func (Mate2pCX) Recombine(rng *rand.Rand, dna1, dna2 DNA) {
	rng = normalizeRNG(rng)
	length := dna1.Len()
	if length < 2 {
		return
	}
	i1, i2 := rng.Intn(length), rng.Intn(length)
	if i1 > i2 {
		i1, i2 = i2, i1
	}
	recombineRange(dna1, dna2, i1, i2)
}

// MateUniformCX swaps genes independently with probability 0.5.
type MateUniformCX struct{}

func (MateUniformCX) Recombine(rng *rand.Rand, dna1, dna2 DNA) {
	rng = normalizeRNG(rng)
	for i := 0; i < dna1.Len(); i++ {
		if rng.Float64() > 0.5 {
			a, b := dna1.Gene(i), dna2.Gene(i)
			dna1.SetGene(i, b)
			dna2.SetGene(i, a)
		}
	}
}

// MateOrderedCX performs ordered crossover for permutation DNA.
type MateOrderedCX struct{}

func (MateOrderedCX) Recombine(rng *rand.Rand, dna1, dna2 DNA) {
	rng = normalizeRNG(rng)
	length := dna1.Len()
	if length < 2 {
		return
	}
	i1, i2 := rng.Intn(length), rng.Intn(length)
	if i1 > i2 {
		i1, i2 = i2, i1
	}
	copy1 := dna1.Copy()
	replaceOrdered(dna1, dna2, i1, i2)
	replaceOrdered(dna2, copy1, i1, i2)
}

func recombineRange(dna1, dna2 DNA, i1, i2 int) {
	for i := i1; i < i2; i++ {
		a, b := dna1.Gene(i), dna2.Gene(i)
		dna1.SetGene(i, b)
		dna2.SetGene(i, a)
	}
}

func replaceOrdered(target, source DNA, i1, i2 int) {
	old := dnaGenes(target)
	newGenes := dnaGenes(source)[i1:i2]
	taken := make(map[int]struct{}, len(newGenes))
	for i, v := range newGenes {
		target.SetGene(i1+i, v)
		taken[int(math.Round(v))] = struct{}{}
	}
	index := 0
	for _, value := range old {
		if _, ok := taken[int(math.Round(value))]; ok {
			continue
		}
		if index == i1 {
			index = i2
		}
		if index >= target.Len() {
			break
		}
		target.SetGene(index, value)
		index++
	}
}

func dnaGenes(dna DNA) []float64 {
	out := make([]float64, dna.Len())
	for i := range out {
		out[i] = dna.Gene(i)
	}
	return out
}

// Selection chooses parents from a candidate population.
type Selection interface {
	Reset(candidates []DNA)
	Pick(rng *rand.Rand, count int) []DNA
}

func fitnessOf(dna DNA) float64 {
	if !dna.HasFitness() {
		return 0
	}
	return dna.Fitness()
}

// RouletteSelection selects candidates proportionally to fitness.
type RouletteSelection struct {
	NegativeValues bool
	candidates     []DNA
	weights        []float64
}

// Reset prepares the selector for a new population.
func (s *RouletteSelection) Reset(candidates []DNA) {
	s.candidates = append([]DNA(nil), candidates...)
	s.weights = make([]float64, len(s.candidates))
	for i, dna := range s.candidates {
		if s.NegativeValues {
			w := fitnessOf(dna)
			if w == 0 {
				s.weights[i] = 0
			} else {
				s.weights[i] = 1 / math.Abs(w)
			}
		} else {
			s.weights[i] = fitnessOf(dna)
		}
	}
}

// Pick returns count selected candidates.
func (s *RouletteSelection) Pick(rng *rand.Rand, count int) []DNA {
	rng = normalizeRNG(rng)
	if len(s.candidates) == 0 {
		return nil
	}
	total := 0.0
	for _, w := range s.weights {
		if w > 0 {
			total += w
		}
	}
	out := make([]DNA, 0, count)
	for len(out) < count {
		var pick DNA
		if total <= 0 {
			pick = s.candidates[rng.Intn(len(s.candidates))]
		} else {
			r := rng.Float64() * total
			acc := 0.0
			pick = s.candidates[len(s.candidates)-1]
			for i, w := range s.weights {
				if w <= 0 {
					continue
				}
				acc += w
				if r < acc {
					pick = s.candidates[i]
					break
				}
			}
		}
		out = append(out, pick)
	}
	return out
}

// RankBasedSelection selects by sorted fitness rank.
type RankBasedSelection struct {
	RouletteSelection
}

// Reset sorts candidates by fitness ascending and assigns rank weights.
func (s *RankBasedSelection) Reset(candidates []DNA) {
	s.candidates = append([]DNA(nil), candidates...)
	sort.SliceStable(s.candidates, func(i, j int) bool {
		return fitnessOf(s.candidates[i]) < fitnessOf(s.candidates[j])
	})
	s.weights = make([]float64, len(s.candidates))
	for i := range s.weights {
		s.weights[i] = float64(i + 1)
	}
}

// TournamentSelection selects the best of a random subset.
type TournamentSelection struct {
	Candidates int
	pool       []DNA
}

// Reset prepares the selector for a new population.
func (s *TournamentSelection) Reset(candidates []DNA) {
	s.pool = append([]DNA(nil), candidates...)
}

// Pick returns count tournament winners.
func (s *TournamentSelection) Pick(rng *rand.Rand, count int) []DNA {
	rng = normalizeRNG(rng)
	if len(s.pool) == 0 {
		return nil
	}
	subset := max(s.Candidates, 1)
	out := make([]DNA, 0, count)
	for len(out) < count {
		best := s.pool[rng.Intn(len(s.pool))]
		for i := 1; i < subset; i++ {
			candidate := s.pool[rng.Intn(len(s.pool))]
			if fitnessOf(candidate) > fitnessOf(best) {
				best = candidate
			}
		}
		out = append(out, best)
	}
	return out
}

// Evaluator computes a fitness value. The optimizer maximizes fitness.
type Evaluator interface {
	Evaluate(dna DNA) float64
}

// EvaluatorFunc adapts a function to Evaluator.
type EvaluatorFunc func(DNA) float64

// Evaluate calls f.
func (f EvaluatorFunc) Evaluate(dna DNA) float64 { return f(dna) }

// LogEntry records one generation.
type LogEntry struct {
	Runtime    time.Duration `json:"runtime"`
	Fitness    float64       `json:"fitness"`
	AvgFitness float64       `json:"avg_fitness"`
}

// Log stores generation statistics.
type Log struct {
	Entries []LogEntry
}

// Add appends an entry.
func (l *Log) Add(runtime time.Duration, fitness, avgFitness float64) {
	l.Entries = append(l.Entries, LogEntry{runtime, fitness, avgFitness})
}

// HallOfFame keeps the best unique fitness values.
type HallOfFame struct {
	Count   int
	entries map[float64]DNA
}

// NewHallOfFame creates a hall of fame with the given capacity.
func NewHallOfFame(count int) *HallOfFame {
	if count < 1 {
		count = 1
	}
	return &HallOfFame{Count: count, entries: make(map[float64]DNA)}
}

// Add records a DNA strand. Fitness must already be set.
func (h *HallOfFame) Add(dna DNA) {
	if !dna.HasFitness() {
		return
	}
	h.entries[dna.Fitness()] = dna
}

// Get returns up to count best entries.
func (h *HallOfFame) Get(count int) []DNA {
	keys := h.sortedKeys()
	if count > h.Count {
		count = h.Count
	}
	if count > len(keys) {
		count = len(keys)
	}
	out := make([]DNA, count)
	for i := range out {
		out[i] = h.entries[keys[i]]
	}
	return out
}

// Purge drops entries beyond the capacity.
func (h *HallOfFame) Purge() {
	keys := h.sortedKeys()
	for i := h.Count; i < len(keys); i++ {
		delete(h.entries, keys[i])
	}
}

func (h *HallOfFame) sortedKeys() []float64 {
	keys := make([]float64, 0, len(h.entries))
	for key := range h.entries {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(keys)))
	return keys
}

// ThresholdFilter removes candidates below a fraction of the best fitness.
func ThresholdFilter(candidates []DNA, bestFitness, threshold float64) []DNA {
	if len(candidates) == 0 {
		return nil
	}
	var minValue float64
	if bestFitness <= 0 {
		minFitness := fitnessOf(candidates[0])
		for _, dna := range candidates[1:] {
			if f := fitnessOf(dna); f < minFitness {
				minFitness = f
			}
		}
		minValue = (1 - threshold) * minFitness
	} else {
		minValue = bestFitness * threshold
	}
	out := make([]DNA, 0, len(candidates))
	for _, dna := range candidates {
		if fitnessOf(dna) > minValue {
			out = append(out, dna)
		}
	}
	return out
}

// Optimizer runs a genetic algorithm.
type Optimizer struct {
	Evaluator Evaluator
	Selection Selection
	Mate      Mate
	Mutation  Mutate
	Rand      *rand.Rand

	MaxGenerations int
	MaxFitness     float64
	MaxRuntime     time.Duration
	MaxStagnation  int
	CrossoverRate  float64
	MutationRate   float64
	Elitism        int
	Threshold      float64

	Candidates   []DNA
	Generation   int
	BestDNA      DNA
	BestFitness  float64
	Stagnation   int
	Log          Log
	HallOfFame   *HallOfFame
	executed     bool
	executedOnce bool
}

// NewOptimizer creates an optimizer with ezdxf-compatible defaults.
func NewOptimizer(evaluator Evaluator, maxGenerations int) *Optimizer {
	if maxGenerations < 1 {
		maxGenerations = 1
	}
	return &Optimizer{
		Evaluator:      evaluator,
		Selection:      &RouletteSelection{},
		Mate:           Mate2pCX{},
		Mutation:       FlipMutate{},
		Rand:           rand.New(rand.NewSource(0)),
		MaxGenerations: maxGenerations,
		MaxFitness:     1.0,
		MaxRuntime:     time.Duration(math.MaxInt64),
		MaxStagnation:  100,
		CrossoverRate:  0.70,
		MutationRate:   0.01,
		Elitism:        2,
		HallOfFame:     NewHallOfFame(10),
	}
}

// ResetFitness sets the initial best fitness. Use a large negative value when
// minimizing by negating the objective.
func (o *Optimizer) ResetFitness(value float64) { o.BestFitness = value }

// AddCandidates appends the initial population. It can only be called before
// Execute.
func (o *Optimizer) AddCandidates(candidates []DNA) error {
	if o.executed {
		return errors.New("genetic: optimizer already executed")
	}
	o.Candidates = append(o.Candidates, candidates...)
	return nil
}

// IsExecuted reports whether Execute has run.
func (o *Optimizer) IsExecuted() bool { return o.executed }

// Count returns the population size.
func (o *Optimizer) Count() int { return len(o.Candidates) }

// Execute runs the optimization loop. feedback is called after at least
// interval has elapsed; returning true stops the loop.
func (o *Optimizer) Execute(feedback func(*Optimizer) bool) error {
	if o.executed {
		return errors.New("genetic: optimizer can only run once")
	}
	if len(o.Candidates) == 0 {
		return errors.New("genetic: no DNA candidates defined")
	}
	if o.Evaluator == nil {
		return errors.New("genetic: evaluator is required")
	}
	if o.Selection == nil {
		o.Selection = &RouletteSelection{}
	}
	if o.Mate == nil {
		o.Mate = Mate2pCX{}
	}
	if o.Mutation == nil {
		o.Mutation = FlipMutate{}
	}
	if o.HallOfFame == nil {
		o.HallOfFame = NewHallOfFame(10)
	}
	if o.Rand == nil {
		o.Rand = rand.New(rand.NewSource(0))
	}
	o.executed = true
	o.executedOnce = true
	start := time.Now()
	lastFeedback := start
	for o.Generation = 1; o.Generation <= o.MaxGenerations; o.Generation++ {
		o.MeasureFitness()
		o.Log.Add(time.Since(start), o.BestFitness, o.averageFitness())
		if o.BestFitness >= o.MaxFitness ||
			time.Since(start) >= o.MaxRuntime ||
			o.Stagnation >= o.MaxStagnation {
			break
		}
		if feedback != nil && time.Since(lastFeedback) >= time.Second {
			if feedback(o) {
				break
			}
			lastFeedback = time.Now()
		}
		o.NextGeneration()
	}
	return nil
}

// MeasureFitness evaluates all unevaluated candidates and updates the best.
func (o *Optimizer) MeasureFitness() {
	o.Stagnation++
	for _, dna := range o.Candidates {
		fitness := fit(dna, o.Evaluator)
		o.HallOfFame.Add(dna)
		if fitness > o.BestFitness {
			o.BestFitness = fitness
			o.BestDNA = dna.Copy()
			o.Stagnation = 0
		}
	}
	o.HallOfFame.Purge()
}

func fit(dna DNA, evaluator Evaluator) float64 {
	if dna.HasFitness() {
		return dna.Fitness()
	}
	fitness := evaluator.Evaluate(dna)
	dna.SetFitness(fitness)
	return fitness
}

func (o *Optimizer) averageFitness() float64 {
	if len(o.Candidates) == 0 {
		return 0
	}
	sum := 0.0
	for _, dna := range o.Candidates {
		sum += fitnessOf(dna)
	}
	return sum / float64(len(o.Candidates))
}

// NextGeneration produces the next population.
func (o *Optimizer) NextGeneration() {
	count := len(o.Candidates)
	if count == 0 {
		return
	}
	candidates := make([]DNA, 0, count)
	pool := o.Candidates
	if o.Threshold > 0 {
		pool = ThresholdFilter(pool, o.BestFitness, o.Threshold)
	}
	if len(pool) == 0 {
		pool = o.Candidates
	}
	o.Selection.Reset(pool)
	if o.Elitism > 0 {
		candidates = append(candidates, o.HallOfFame.Get(o.Elitism)...)
	}
	for len(candidates) < count {
		parents := o.Selection.Pick(o.Rand, 2)
		if len(parents) == 0 {
			break
		}
		dna1 := parents[0].Copy()
		dna2 := parents[len(parents)-1].Copy()
		if dna1.Len() != 0 {
			o.recombine(dna1, dna2)
			o.mutate(dna1, dna2)
		}
		candidates = append(candidates, dna1)
		if len(candidates) < count {
			candidates = append(candidates, dna2)
		}
	}
	o.Candidates = candidates
}

func (o *Optimizer) recombine(a, b DNA) {
	if o.Rand.Float64() < o.CrossoverRate {
		o.Mate.Recombine(o.Rand, a, b)
	}
}

func (o *Optimizer) mutate(a, b DNA) {
	o.Mutation.Mutate(o.Rand, a, o.MutationRate)
	o.Mutation.Mutate(o.Rand, b, o.MutationRate)
}

// String describes the optimizer state.
func (o *Optimizer) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "GeneticOptimizer(generation=%d, candidates=%d, best=%.6f", o.Generation, len(o.Candidates), o.BestFitness)
	if o.BestDNA != nil {
		fmt.Fprintf(&b, ", dna=%s", o.BestDNA.String())
	}
	b.WriteString(")")
	return b.String()
}

// Capability registers the add-on with the addons registry.
type Capability struct{}

// Name implements addons.Capability.
func (Capability) Name() string { return "genetic_algorithm" }

// Description implements addons.Capability.
func (Capability) Description() string { return "Genetic algorithm optimization helpers" }

func init() {
	_ = addons.Register(Capability{}.Name(), Capability{})
}
