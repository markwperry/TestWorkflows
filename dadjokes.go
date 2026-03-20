package main

import "math/rand"

type dadJoke struct {
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

var dadJokes = []dadJoke{
	{"Why do programmers prefer dark mode?", "Because light attracts bugs."},
	{"Why do Java developers wear glasses?", "Because they can't C#."},
	{"What's a programmer's favorite hangout place?", "Foo Bar."},
	{"Why was the JavaScript developer sad?", "Because he didn't Node how to Express himself."},
	{"How many programmers does it take to change a light bulb?", "None. That's a hardware problem."},
	{"Why do programmers always mix up Halloween and Christmas?", "Because Oct 31 == Dec 25."},
	{"What's the object-oriented way to become wealthy?", "Inheritance."},
	{"Why did the developer go broke?", "Because he used up all his cache."},
	{"What do you call a programmer from Finland?", "Nerdic."},
	{"Why did the functions stop calling each other?", "Because they got too many arguments."},
	{"What's a bug's least favorite thing?", "A debugger."},
	{"Why do Python programmers have low self-esteem?", "They're constantly comparing themselves to others."},
}

func getRandomDadJoke() dadJoke {
	return dadJokes[rand.Intn(len(dadJokes))]
}
