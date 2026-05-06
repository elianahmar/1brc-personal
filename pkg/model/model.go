package model

type City string

type Validation struct {
	Min float64
	Max float64
	Avg float64
}

type Measurement struct {
	City  string
	Temps float64
	Count int
	Min   float64
	Max   float64
	Avg   float64
}
