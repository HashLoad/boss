package crypto_test

import (
	"testing"

	"github.com/hashload/boss/utils/crypto"
)

func TestCryptoa(t *testing.T) {
	type args struct {
		key     []byte
		message string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test 1",
			args: args{
				key:     []byte("1234567890123456"),
				message: "Hello, World!",
			}},
		{
			name: "Test 2",
			args: args{
				key:     []byte("1234567890123456"),
				message: "Hello, World!",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crypto.Encrypt(tt.args.key, tt.args.message)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			dec, err := crypto.Decrypt(tt.args.key, got)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			if dec != tt.args.message {
				t.Errorf("Decrypt() = %v, want %v", dec, tt.args.message)
			}
		})
	}
}
