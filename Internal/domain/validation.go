package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrOrderUidEmpty   = errors.New("order_uid is required")
	ErrOrderUidInvalid = errors.New("order_uid has invalid format")
	ErrOrderInvalid    = errors.New("order validation failed")
)

var orderUidRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]{1,128}$`)

func ValidateOrderUid(orderUid string) error {
	s := strings.TrimSpace(orderUid)
	if s == "" {
		return ErrOrderUidEmpty
	}
	if !orderUidRegex.MatchString(s) {
		return ErrOrderUidInvalid
	}
	return nil
}

func ValidateOrder(o *Order) error {
	if o == nil {
		return ErrOrderInvalid
	}
	if strings.TrimSpace(o.Orders.OrderUid) == "" {
		return ErrOrderUidEmpty
	}
	if !orderUidRegex.MatchString(o.Orders.OrderUid) {
		return ErrOrderUidInvalid
	}
	if len(o.Items) == 0 {
	}
	return nil
}
