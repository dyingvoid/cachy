package features_replication

type Response interface {
	isResponse()
}

type PongResponse struct{}

func (PongResponse) isResponse() {}

type OKResponse struct{}

func (OKResponse) isResponse() {}

type PsyncResponse struct {
	ReplID string
	Offset int64
}

func (PsyncResponse) isResponse() {}
