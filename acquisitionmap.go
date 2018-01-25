package main

// AcquisitionMap keeps track of the number of firmware versions acquired for
// each model.
type AcquisitionMap struct {
	required int
	m        map[string]int
}

// Require updates the required number of firmware versions for each model.
func (a *AcquisitionMap) Require(count int) {
	a.required = count
}

// Match returns true if the required number of firmware versions for the model
// have been acquired.
func (a *AcquisitionMap) Match(model string) bool {
	if a.required == 0 {
		return false
	}
	return a.m[model] >= a.required
}

// Add increments the counter for the given model.
func (a *AcquisitionMap) Add(models ...string) {
	if a.m == nil {
		a.m = make(map[string]int)
	}
	for _, model := range models {
		a.m[model]++
	}
}
