package models

// Stats represents cache statistics information such as hits, misses, evictions, total keys, and memory usage.
type Stats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	KeysTotal int
}
