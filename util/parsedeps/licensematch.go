// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"regexp"
	"sort"

	"github.com/agnivade/levenshtein"
)

// Ideally we'd use github.com/google/licenseclassifier, which is much
// more extensive than this code, but it has dependencies on go1.16+, and
// we're still using go1.15. We don't want to have all users maintain a
// separate go version just for this.

type LicMatch struct {
	Name          string
	Distance      int
	DistanceRatio float32
}

var reJunk = regexp.MustCompile(`[\*=-]+`)
var reSpace = regexp.MustCompile(`\s+`)

type LicenseMatcher struct {
	licenses []License
}

func NewLicenseMatcher() *LicenseMatcher {
	matcher := LicenseMatcher{
		licenses: []License{},
	}
	for _, lic := range Licenses {
		licCopy := License{
			Name: lic.Name,
			Text: reSpace.ReplaceAllString(lic.Text, " "),
		}
		licCopy.Text = reJunk.ReplaceAllString(licCopy.Text, "")
		matcher.licenses = append(matcher.licenses, licCopy)
	}
	return &matcher
}

func (s *LicenseMatcher) Match(contents string) (string, float32) {
	contents = reSpace.ReplaceAllString(contents, " ")
	contents = reJunk.ReplaceAllString(contents, "")
	matches := []LicMatch{}
	for _, lic := range s.licenses {
		dist := levenshtein.ComputeDistance(contents, lic.Text)
		// The distance is the number of changes to get from
		// one string to the other. Take the ratio of this against
		// the average number of characters in the text to get
		// a better number to compare against, because if two small
		// files require the same number of changes to match as
		// two large files, the two large files have more matching
		// content.
		distRatio := float32(dist) / float32(len(contents)+len(lic.Text)) / 2.0
		match := LicMatch{
			Name:          lic.Name,
			Distance:      dist,
			DistanceRatio: distRatio,
		}
		matches = append(matches, match)
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].DistanceRatio < matches[j].DistanceRatio
	})
	if matches[0].DistanceRatio > 0.1 {
		// more than 10% of the file is different, it probably is not
		// a real match
		return UnknownLicense, matches[0].DistanceRatio
	}
	return matches[0].Name, matches[0].DistanceRatio
}
