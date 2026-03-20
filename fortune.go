package main

import (
	"fmt"
	"math/rand"
)

var fortunes = []string{
	"A beautiful, smart, and loving person will be coming into your life.",
	"A dubious friend may be an enemy in camouflage.",
	"A faithful friend is a strong defense.",
	"Your code will compile on the first try... just kidding.",
	"A merge conflict is in your near future.",
	"The bug you seek is three lines above where you're looking.",
	"You will find the answer on Stack Overflow, but only after posting your question.",
	"Today's commit message will be 'fixed stuff'.",
	"The tests you skip today will haunt you tomorrow.",
	"A great software developer is not born, but compiled through experience.",
	"The cloud is just someone else's computer having a bad day.",
	"You will mass-refactor and regret nothing... for now.",
}

var luckyNumbers = []int{4, 8, 15, 16, 23, 42, 69, 420, 1337}

func getRandomFortune() (string, []int) {
	fortune := fortunes[rand.Intn(len(fortunes))]
	nums := make([]int, 3)
	for i := range nums {
		nums[i] = luckyNumbers[rand.Intn(len(luckyNumbers))]
	}
	return fmt.Sprintf("%s Lucky numbers: %d, %d, %d", fortune, nums[0], nums[1], nums[2]), nums
}
