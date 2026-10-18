package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 19 * 1024
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
)

var hashSlots = make(chan struct{}, max(2, runtime.GOMAXPROCS(0)))

func acquire(ctx context.Context) (func(), error) {
	select {
	case hashSlots <- struct{}{}:
		return func() { <-hashSlots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func VerifyPassword(password, encoded string) bool {
	var salt, want []byte
	var t, m uint32
	var p uint8
	var err error
	if strings.HasPrefix(encoded, "$argon2id$") {
		parts := strings.Split(encoded, "$")
		if len(parts) != 6 {
			return false
		}
		if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
			return false
		}
		if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
			return false
		}
		if want, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
			return false
		}
	} else {
		parts := strings.Split(encoded, "$")
		if len(parts) != 2 {
			return false
		}
		if salt, err = base64.RawStdEncoding.DecodeString(parts[0]); err != nil {
			return false
		}
		if want, err = base64.RawStdEncoding.DecodeString(parts[1]); err != nil {
			return false
		}
		t, m, p = 1, 64*1024, 4
	}
	if m > 256*1024 || t > 10 || p == 0 || len(want) == 0 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func NeedsRehash(encoded string) bool {
	return !strings.HasPrefix(encoded, fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$", argon2.Version, argonMemory, argonTime, argonThreads))
}
