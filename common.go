package provicol

type askingBytecode uint8

const (
	Ping askingBytecode = iota
	Connect
	ListBuckets
	ListObjects
	GetObject
	PutObject
	CreateBucket
	GetInfos
)

const MAGIC_NUMBER uint32 = 0x1b505643
