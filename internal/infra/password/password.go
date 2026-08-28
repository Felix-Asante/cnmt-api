package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"cnmt/internal/common/env"

	"golang.org/x/crypto/argon2"
)

const (
	defaultTime    = 3
	defaultMemory  = 64 * 1024
	defaultThreads = 2
	defaultKeyLen  = 32
	defaultSaltLen = 16
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLen      uint32
}

func Hash(plain string) (string, error) {
	p := loadParams()

	salt := make([]byte, defaultSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(plain), salt, p.iterations, p.memory, p.parallelism, p.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

func Compare(encoded, plain string) bool {
	p, salt, hash, err := decode(encoded)
	if err != nil {
		return false
	}

	other := argon2.IDKey([]byte(plain), salt, p.iterations, p.memory, p.parallelism, p.keyLen)
	return subtle.ConstantTimeCompare(hash, other) == 1
}

func loadParams() params {
	return params{
		memory:      uint32(env.GetInt("ARGON2_MEMORY_KB", defaultMemory)),
		iterations:  uint32(env.GetInt("ARGON2_ITERATIONS", defaultTime)),
		parallelism: uint8(env.GetInt("ARGON2_PARALLELISM", defaultThreads)),
		keyLen:      defaultKeyLen,
	}
}

func decode(encoded string) (params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return params{}, nil, nil, fmt.Errorf("invalid argon2 hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return params{}, nil, nil, err
	}
	if version != argon2.Version {
		return params{}, nil, nil, fmt.Errorf("unsupported argon2 version")
	}

	p := params{keyLen: defaultKeyLen}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism); err != nil {
		return params{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params{}, nil, nil, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params{}, nil, nil, err
	}

	p.keyLen = uint32(len(hash))
	return p, salt, hash, nil
}