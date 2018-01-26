package main

import (
	"fmt"
	"strings"

	fw "github.com/scjalliance/dpmafirmware"
)

// CacheStatus records the cache status for a set of models.
type CacheStatus struct {
	Failed  fw.ModelSet
	Missing fw.ModelSet
	Revised fw.ModelSet
	Current fw.ModelSet
	md5Sum  string
}

// Good returns the number of models with current cache entries.
func (status *CacheStatus) Good() int {
	return len(status.Current)
}

// Bad returns the number of models with failed, missing or out of date cache
// entries.
func (status *CacheStatus) Bad() int {
	return len(status.Failed) + len(status.Missing) + len(status.Revised)
}

// Needed returns the models in need of updating.
func (status *CacheStatus) Needed() (needed fw.ModelSet) {
	needed = make(fw.ModelSet, 0, status.Bad())
	needed = append(needed, status.Revised...)
	needed = append(needed, status.Missing...)
	needed = append(needed, status.Failed...)
	return
}

// Summary returns a summary of the cache status.
func (status *CacheStatus) Summary() []string {
	var output []string
	add := func(kind string, models fw.ModelSet) {
		if len(models) > 0 {
			output = append(output, fmt.Sprintf("[%s]: %-11s  md5: %s", strings.Join(models, ","), kind, status.md5Sum))
		}
	}
	add("Up to date", status.Current)
	add("Out of date", status.Revised)
	add("Missing", status.Missing)
	add("Cache error", status.Failed)
	return output
}
