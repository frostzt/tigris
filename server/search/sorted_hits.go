// Copyright 2022-2023 Tigris Data, Inc.
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

package search

import (
	"encoding/json"

	"github.com/tigrisdata/tigris/lib/container"
	"github.com/tigrisdata/tigris/query/sort"
	ulog "github.com/tigrisdata/tigris/util/log"
	tsApi "github.com/typesense/typesense-go/typesense/api"
)

type Hit struct {
	Document       map[string]interface{}
	TextMatchScore int64
}

// True - field absent in document
// True - field values is nil
// False - field has non-nil value.
func (sh *Hit) isFieldMissingOrNil(f string) bool {
	if v, ok := sh.Document[f]; !ok {
		return true
	} else {
		return v == nil
	}
}

func NewSearchHit(tsHit *tsApi.SearchResultHit) *Hit {
	if tsHit == nil || tsHit.Document == nil {
		return nil
	}
	score := int64(0)
	if tsHit.TextMatch != nil {
		score = *tsHit.TextMatch
	}
	return &Hit{
		Document:       *tsHit.Document,
		TextMatchScore: score,
	}
}

// SortedMergeHits is a Priority queue to merge sorted results. This is used when we have multiple responses and we know
// that each response hits is already sorted.
type SortedMergeHits struct {
	hits *container.PriorityQueue[Hit]
}

// NewSortedHits returns hits in decreasing order by priority.
func NewSortedHits(sortingOrder *sort.Ordering) *SortedMergeHits {
	s := SortedMergeHits{
		hits: container.NewPriorityQueue[Hit](func(this, that *Hit) bool {
			return hitsComparator(this, that, sortingOrder) > 0
		}),
	}

	return &s
}

func (s *SortedMergeHits) add(hit *Hit) {
	if hit != nil || hit.Document != nil {
		s.hits.Push(hit)
	}
}

func (s *SortedMergeHits) Next() (*Hit, error) {
	return s.hits.Pop()
}

func (s *SortedMergeHits) Len() int {
	return s.hits.Len()
}

func (s *SortedMergeHits) HasMoreHits() bool {
	return s.Len() > 0
}

// Comparison outputs.
const (
	This  = 1
	That  = -1
	Equal = 0
)

// hitsComparator returns 0 if priority of this and that are equal
// return positive int if priority of "this" is greater than "that"
// return negative int if priority of "that" is greater than "this".
func hitsComparator(this, that *Hit, sortingOrder *sort.Ordering) int {
	// equal priority; although nil pointers are not expected
	if this == nil && that == nil {
		return Equal
	}
	// safety; nil pointers are not expected
	if that == nil {
		return This
	} else if this == nil {
		return That
	}

	// only consider 2 sorting orders for now
	for i := 0; i < 2 && sortingOrder != nil; i++ {
		if i >= len(*sortingOrder) {
			break
		}
		order := (*sortingOrder)[i]

		thisIsNil, thatIsNil := this.isFieldMissingOrNil(order.Name), that.isFieldMissingOrNil(order.Name)

		// if both hits are missing/nil field, continue
		if thisIsNil && thatIsNil {
			continue
		}

		// only one of the document is missing the field
		if (thisIsNil && order.MissingValuesFirst) || (thatIsNil && !order.MissingValuesFirst) {
			return This
		} else if (thatIsNil && order.MissingValuesFirst) || (thisIsNil && !order.MissingValuesFirst) {
			return That
		}

		// extract values to perform actual comparison
		thisVal, thatVal := this.Document[order.Name], that.Document[order.Name]
		var thisV, thatV float64

		switch v := thisVal.(type) {
		case json.Number:
			var err error
			thisV, err = v.Float64()
			// log the number conversion error and continue to next comparison
			if ulog.E(err) {
				continue
			}
			thatV, err = thatVal.(json.Number).Float64()
			if ulog.E(err) {
				continue
			}
		case bool:
			if v {
				thisV = 1
			}
			if thatVal.(bool) {
				thatV = 1
			}
		default:
			// String or other comparisons not supported at the moment,
			// also not expected to receive any unexpected field types here, just skip
			continue
		}

		// if values are equal, eval next sort condition
		if thisV == thatV {
			continue
		}

		if (thisV > thatV && order.Ascending) || (thatV > thisV && !order.Ascending) {
			return That
		} else if (thisV > thatV && !order.Ascending) || (thatV > thisV && order.Ascending) {
			return This
		}
	}

	// break the tie using highest TextMatch score to appear first when using Pop() operation
	if this.TextMatchScore > that.TextMatchScore {
		return This
	} else if this.TextMatchScore < that.TextMatchScore {
		return That
	}
	return Equal
}
