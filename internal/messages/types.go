package messages

var (
	TypePushPayment        = byte(1)
	TypeSummarizePayments  = byte(2)
	TypeSummarizedPayments = byte(3)
)

type Payment struct {
	Amount      float64 `json:"amount" msgpack:"amount"`
	CID         string  `json:"correlationId" msgpack:"correlationId"`
	RequestedAt string  `json:"requestedAt" msgpack:"requestedAt"`
}

type SummarizePayments struct {
	From string `msgpack:"from"`
	To   string `msgpack:"to"`
}

type SummarizedPayments struct {
	Default  SummarizedProcessor `msgpack:"default" json:"default"`
	Fallback SummarizedProcessor `msgpack:"fallback" json:"fallback"`
}

type SummarizedProcessor struct {
	TotalAmount   float64 `msgpack:"totalAmount" json:"totalAmount"`
	TotalRequests int64   `msgpack:"totalRequests" json:"totalRequests"`
}
