package types

type Payment struct {
	Amount      float64 `json:"amount" msgpack:"amount"`
	CID         string  `json:"correlationId" msgpack:"correlationId"`
	RequestedAt string  `json:"requestedAt" msgpack:"requestedAt"`
}
