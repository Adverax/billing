package service

type CreditRequest struct {
	Uid     int64
	Account uint32
	Amount  float32
}

type CreditResponse struct {
	Status uint8
}

type DebitRequest struct {
	Uid     int64
	Account uint32
	Amount  float32
}

type DebitResponse struct {
	Status uint8
}

type TransferRequest struct {
	Uid    int64
	Src    uint32
	Dst    uint32
	Amount float32
}

type TransferResponse struct {
	Status uint8
}

type AcquireRequest struct {
	Uid     int64
	Account uint32
	Amount  float32
}

type AcquireResponse struct {
	Status uint8
}

type CommitRequest struct {
	Uid     int64
	Account uint32
}

type CommitResponse struct {
	Status uint8
}

type RollbackRequest struct {
	Uid     int64
	Account uint32
}

type RollbackResponse struct {
	Status uint8
}
