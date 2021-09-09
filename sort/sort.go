package sort

import (
	"fmt"
	"sort"
)

type kv struct {
	Key   string
	Value float64
}

func (ss kv) String() string {
	return fmt.Sprintf("%v - %v", ss.Key, ss.Value)
}

func Sort(data map[string]float64) []kv {

	var ss []kv
	for k, v := range data {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss

}
