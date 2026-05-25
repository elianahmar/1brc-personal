package model

import "fmt"

type Compute interface {
	Compute() map[string]*Measurement
}
type Validation struct {
	City string
	Min  float64
	Max  float64
	Avg  float64
}

type MeasurementInt struct {
	City     string
	Temps    int
	TempsInt int
	Count    int
	Min      int
	Max      int
	Avg      float64
}

func (p *Predicted) Print() {
	println(fmt.Sprintf("City = %s, min/max/avg = %s/%s/%s", p.City, p.Min, p.Max, p.Avg))
}

type Actual struct {
	City string
	Min  string
	Max  string
	Avg  string
}

type Predicted struct {
	City string
	Min  string
	Max  string
	Avg  string
}

type Result struct {
	City string
	Min  float64
	Max  float64
	Avg  float64
}

type Measurement struct {
	City     string
	Temps    float64
	TempsInt int
	Count    float64
	Min      float64
	Max      float64
	Avg      float64
}

type ReadChunk struct {
	Buffer []byte
	Offset int
	Idx    int
}

type Chunk struct {
	BufSize int
	Offset  int
	Idx     int
}

type Line struct {
	// What chunk it appears in
	ChunkIdx int
	// Full line as byte slice
	Line []byte
	// Index of the line after we split the bytes on '\n'
	LineIdx int
}

type Range struct {
	Start int64
	End   int64
	Index int
}
