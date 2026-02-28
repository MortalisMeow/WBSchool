package domain

import (
	"testing"
)

func repeatByte(b byte, n int) []byte {
	s := make([]byte, n)
	for i := range s {
		s[i] = b
	}
	return s
}

func TestValidateOrderUid(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		wantErr bool
	}{
		{"empty", "", true},
		{"space only", "   ", true},
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid short", "abc-123", false},
		{"invalid chars", "order/../x", true},
		{"invalid space", "order uid", true},
		{"too long", string(repeatByte('a', 129)), true},
		{"valid 128", string(repeatByte('a', 128)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderUid(tt.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOrderUid(%q) err = %v, wantErr %v", tt.uid, err, tt.wantErr)
			}
		})
	}
}

func TestValidateOrder(t *testing.T) {
	t.Run("nil order", func(t *testing.T) {
		if err := ValidateOrder(nil); err == nil {
			t.Error("ValidateOrder(nil) expected error")
		}
	})
	t.Run("empty order_uid", func(t *testing.T) {
		o := &Order{Orders: Orders{OrderUid: ""}}
		if err := ValidateOrder(o); err != ErrOrderUidEmpty {
			t.Errorf("ValidateOrder(empty uid) = %v, want ErrOrderUidEmpty", err)
		}
	})
	t.Run("invalid order_uid format", func(t *testing.T) {
		o := &Order{Orders: Orders{OrderUid: "bad uid"}}
		if err := ValidateOrder(o); err != ErrOrderUidInvalid {
			t.Errorf("ValidateOrder(bad uid) = %v, want ErrOrderUidInvalid", err)
		}
	})
	t.Run("valid order", func(t *testing.T) {
		o := &Order{Orders: Orders{OrderUid: "550e8400-e29b-41d4-a716-446655440000"}}
		if err := ValidateOrder(o); err != nil {
			t.Errorf("ValidateOrder(valid) = %v", err)
		}
	})
}
