package password

import (
	"strings"
	"testing"
)

func TestHash(t *testing.T) {
	password := "correcthorsebatterystaple"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	// Check PHC format
	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Errorf("Hash() format invalid, got %s", hash)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Hash() expected 6 parts, got %d", len(parts))
	}
}

func TestVerify(t *testing.T) {
	password := "mysecretpassword"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  error
	}{
		{
			name:     "correct password",
			hash:     hash,
			password: password,
			wantErr:  nil,
		},
		{
			name:     "wrong password",
			hash:     hash,
			password: "wrongpassword",
			wantErr:  ErrMismatch,
		},
		{
			name:     "invalid hash format",
			hash:     "notahash",
			password: password,
			wantErr:  ErrInvalidHash,
		},
		{
			name:     "empty password against valid hash",
			hash:     hash,
			password: "",
			wantErr:  ErrMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.hash, tt.password)
			if err != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashUniqueness(t *testing.T) {
	password := "samepassword"

	hash1, _ := Hash(password)
	hash2, _ := Hash(password)

	if hash1 == hash2 {
		t.Error("Hash() should produce unique hashes for same password (different salts)")
	}

	// Both should still verify
	if err := Verify(hash1, password); err != nil {
		t.Errorf("hash1 verification failed: %v", err)
	}
	if err := Verify(hash2, password); err != nil {
		t.Errorf("hash2 verification failed: %v", err)
	}
}

func TestHashWithParams(t *testing.T) {
	password := "testpassword"

	params := &Params{
		Memory:      32 * 1024,
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := HashWithParams(password, params)
	if err != nil {
		t.Fatalf("HashWithParams() error = %v", err)
	}

	// Should contain the custom params
	if !strings.Contains(hash, "m=32768,t=2,p=1") {
		t.Errorf("HashWithParams() params not encoded correctly: %s", hash)
	}

	// Should still verify
	if err := Verify(hash, password); err != nil {
		t.Errorf("Verify() failed for custom params: %v", err)
	}
}

func TestNeedsRehash(t *testing.T) {
	password := "testpassword"

	// Hash with default params - should not need rehash
	hash, _ := Hash(password)
	if NeedsRehash(hash) {
		t.Error("NeedsRehash() should return false for default params")
	}

	// Hash with different params - should need rehash
	oldParams := &Params{
		Memory:      32 * 1024, // Different from default
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
	oldHash, _ := HashWithParams(password, oldParams)
	if !NeedsRehash(oldHash) {
		t.Error("NeedsRehash() should return true for non-default params")
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr error
	}{
		{
			name:    "empty string",
			hash:    "",
			wantErr: ErrInvalidHash,
		},
		{
			name:    "random string",
			hash:    "randomgarbage",
			wantErr: ErrInvalidHash,
		},
		{
			name:    "wrong algorithm",
			hash:    "$argon2i$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$c29tZWhhc2g",
			wantErr: ErrInvalidHash,
		},
		{
			name:    "malformed params",
			hash:    "$argon2id$v=19$invalid$c29tZXNhbHQ$c29tZWhhc2g",
			wantErr: ErrInvalidHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.hash, "anypassword")
			if err != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{"default length (0)", 0, 16},
		{"custom length 8", 8, 8},
		{"custom length 32", 32, 32},
		{"negative length", -5, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.length)
			if len(got) != tt.want {
				t.Errorf("Generate(%d) length = %d, want %d", tt.length, len(got), tt.want)
			}
		})
	}

	// Test uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p := Generate(16)
		if seen[p] {
			t.Error("Generate() produced duplicate password")
		}
		seen[p] = true
	}
}

func TestMatch(t *testing.T) {
	password := "testpassword"
	hash, _ := Hash(password)

	if !Match(hash, password) {
		t.Error("Match() = false, want true for correct password")
	}

	if Match(hash, "wrongpassword") {
		t.Error("Match() = true, want false for wrong password")
	}

	if Match("invalidhash", password) {
		t.Error("Match() = true, want false for invalid hash")
	}
}

func BenchmarkHash(b *testing.B) {
	password := "benchmarkpassword"
	for i := 0; i < b.N; i++ {
		Hash(password)
	}
}

func BenchmarkVerify(b *testing.B) {
	password := "benchmarkpassword"
	hash, _ := Hash(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(hash, password)
	}
}
