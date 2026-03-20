package main

import "math/rand"

func flipCoin() string {
	if rand.Intn(2) == 0 {
		return "heads"
	}
	return "tails"
}

func flipMultiple(n int) map[string]int {
	results := map[string]int{"heads": 0, "tails": 0}
	for range n {
		results[flipCoin()]++
	}
	return results
}
