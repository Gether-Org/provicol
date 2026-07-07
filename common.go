package provicol

type askingBytecode uint8

const (
	Ping    askingBytecode = iota // ( ) -> ( string )
	Connect                       // ( credentials [ ] byte ) -> ( string )

	ListBuckets  // ( pukey string ) -> ( [ ] string )
	CreateBucket // ( pukey string, bucket string ) -> ( )
	DeleteBucket // ( pukey string, bucket string ) -> ( )

	ListObjects    // ( pukey string, bucket string ) -> ( [ ] string )
	GetObject      // ( pukey string, bucket string, object string ) -> ( [ ] byte )
	PutObject      // ( pukey string, bucket string, object string, data [ ] byte ) -> ( )
	DeleteObject   // ( pukey string, bucket string, object string ) -> ( )
	GetObjectInfos // ( puket string, bucket string, object string ) -> ( date string, size int64, author string )
)

const (
	// date used for layour of GetObjectInfos
	provicolDateLayout        = "2006-01-02 15:04:05"
	magicNumber        uint32 = 0x1b505643
)
