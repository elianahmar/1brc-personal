package model

type City string

type Compute interface {
	Compute() map[City]*Measurement
}
type Validation struct {
	City string
	Min  float64
	Max  float64
	Avg  float64
}

type Measurement struct {
	City  City
	Temps float64
	Count float64
	Min   float64
	Max   float64
	Avg   float64
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
