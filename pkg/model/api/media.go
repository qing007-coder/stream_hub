package api

type InitUploadReq struct {
	FileSize int64  `json:"file_size"`
	FileHash string `json:"file_hash"`
	FileType string `json:"file_type"`
}

type InitUploadResp struct {
	IsSkipped   bool   `json:"is_skipped"`
	VideoURL    string `json:"video_url"`
	UploadID    string `json:"upload_id"`
	FinishChunk []int  `json:"finish_chunk"`
	ChunkSize   int64  `json:"chunk_size"`
}

type UploadChunkResp struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"e_tag"`
	Size       int64  `json:"size"`
}

type CompleteUploadReq struct {
	FileHash string `json:"file_hash"`
	UploadID string `json:"upload_id"`
}

type CompleteUploadResp struct {
	VideoURL string `json:"video_url"`
}
