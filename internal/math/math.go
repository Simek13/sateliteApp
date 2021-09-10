package math

import "time"

func Min(nums []float64) (min float64) {
	min = nums[0]
	for _, num := range nums {
		if num < min {
			min = num
		}
	}
	return
}

func Max(nums []float64) (max float64) {
	max = nums[0]
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return
}

func Avg(nums []float64) (avg float64) {
	total := .0
	for _, num := range nums {
		total += num
	}
	avg = total / float64(len(nums))
	return
}

func MinDate(dates []time.Time) (min time.Time) {
	min = dates[0]
	for _, date := range dates {
		if min.After(date) {
			min = date
		}
	}
	return
}

func MaxDate(dates []time.Time) (max time.Time) {
	max = dates[0]
	for _, date := range dates {
		if max.Before(date) {
			max = date
		}
	}
	return
}
