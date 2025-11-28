package provicol

type askingBytecode uint8

const (
	Ping askingBytecode = iota
	Connect
	ListBuckets
	ListObjects
	GetObject
)
