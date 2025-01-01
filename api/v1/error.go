package v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type ErrorOffsetOutOfRange struct {
	Offset uint64
}

func (e *ErrorOffsetOutOfRange) GRPCStatus() *status.Status {
	st := status.New(
		404,
		fmt.Sprintf("offset out of range: %d", e.Offset),
	)

	msg := fmt.Sprintf("record at offset %d is outside log range", e.Offset)
	details := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: msg,
	}

	str, err := st.WithDetails(details)
	if err != nil {
		return st
	}

	return str

}

func (e *ErrorOffsetOutOfRange) Error() string {
	return e.GRPCStatus().Message()
}
