package provicol

type askingBytecode uint8

const (
	Ping    askingBytecode = iota // ( ) -> ( string )
	Connect                       // ( credentials [ ] byte ) -> ( string )

	// Buckets

	ListBuckets  // ( pukey string ) -> ( [ ] string )
	CreateBucket // ( pukey string, bucket string ) -> ( )
	DeleteBucket // ( pukey string, bucket string ) -> ( )

	ListObjects    // ( pukey string, bucket string ) -> ( [ ] string )
	GetObject      // ( pukey string, bucket string, object string ) -> ( data []byte or file descriptor )
	PutObject      // ( pukey string, bucket string, object string, data []byte or file descriptor ) -> ( )
	DeleteObject   // ( pukey string, bucket string, object string ) -> ( )
	GetObjectInfos // ( puket string, bucket string, object string ) -> ( date string, size int64, author string )

	// Registery

	ListRepositories // ( id ) -> ( []string )
	CreateRepository // ( id, repo ) -> ( )
	DeleteRepository // ( id, repo ) -> ( )
	ListTags         // ( id, repo ) -> ( []string )

	GetManifest    // ( id, repo, ref ) -> ( mediaType string, digest string, data []byte )
	PutManifest    // ( id, repo, ref, mediaType string, data []byte ) -> ( )
	DeleteManifest // ( id, repo, ref ) -> ( )

	StatBlob        // ( id, repo, digest ) -> ( exists bool, size int64 )
	GetBlobChunk    // ( id, repo, digest, offset, length int64 ) -> ( []byte )
	BeginBlobUpload // ( id, repo ) -> ( uploadID string )
	PutBlobChunk    // ( id, uploadID, offset int64, data []byte ) -> ( )
	CommitBlob      // ( id, uploadID, digest string ) -> ( )

	Capabilities // ( ) -> ( []string )   // "storage", "registry"
)

const (
	// date used for layour of GetObjectInfos
	ProvicolDateLayout        = "2006-01-02 15:04:05"
	magicNumber        uint32 = 0x1b505643
)
