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

// order_uid: допустимы UUID, буквы, цифры, дефис (не пустой, разумная длина)
var orderUidRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]{1,128}$`)

// ValidateOrderUid проверяет параметр order_uid из запроса (path/query).
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

// ValidateOrder проверяет структуру заказа после десериализации.
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
	// Минимальная проверка вложенных структур
	if len(o.Items) == 0 {
		// допускаем 0 items по бизнес-логике, но можно ужесточить
	}
	return nil
}
