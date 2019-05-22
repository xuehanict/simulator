package mara

import "fmt"

const (
	ALLOCATION_NOT_MATCH_ROUTE  =  iota
	FIND_PATH_FAILED
	ALLOCARION_FAILED
	UPDATE_LINK_FAILED
)

type PaymentError struct {
	Code int
	Description string
}

func (e *PaymentError) Error() string {
	return fmt.Sprintf("error code %d : %s", e.Code, e.Description)
}

