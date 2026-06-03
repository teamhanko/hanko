package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"regexp"
	"strconv"
	"strings"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"golang.org/x/crypto/bcrypt"

	"golang.org/x/crypto/scrypt"
)

var (
	ErrorPasswordInvalid = errors.New("password invalid")
)

type Password interface {
	VerifyPassword(tx *pop.Connection, userId uuid.UUID, password string) error
	RecoverPassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error
	CreatePassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error
	UpdatePassword(tx *pop.Connection, passwordCredentialModel *models.PasswordCredential, newPassword string) error
}

type password struct {
	persister persistence.Persister
}

var fbscryptHashRegexp = regexp.MustCompile(`^\$fbscrypt\$v=(?P<v>[0-9]+),n=(?P<n>[0-9]+),r=(?P<r>[0-9]+),p=(?P<p>[0-9]+)(?:,ss=(?P<ss>[^,]+))?(?:,sk=(?P<sk>[^$]+))?\$(?P<salt>[^$]+)\$(?P<hash>.+)$`)

const (
	FirebaseScryptPrefix = "$fbscrypt"
	FirebaseScryptKeyLen = 32
)

type FirebaseScryptHashInput struct {
	v             string
	memCost       uint64
	rounds        uint64
	parallelism   uint64
	saltSeparator []byte
	signerKey     []byte
	salt          []byte
	rawHash       []byte
}

func NewPasswordService(persister persistence.Persister) Password {
	return &password{
		persister,
	}
}

func (s password) VerifyPassword(tx *pop.Connection, userId uuid.UUID, password string) error {
	user, err := s.persister.GetUserPersisterWithConnection(tx).Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return ErrorPasswordInvalid
	}

	pw, err := s.persister.GetPasswordCredentialPersisterWithConnection(tx).GetByUserID(userId)
	if err != nil {
		return fmt.Errorf("error retrieving password credential: %w", err)
	}

	if pw == nil {
		return ErrorPasswordInvalid
	}

	if err = s.CompareHashAndPassword(pw.Password, password); err != nil {
		return err
	}

	return nil
}

func (s password) CompareHashAndPassword(hash, password string) error {
	if strings.HasPrefix(hash, FirebaseScryptPrefix) {
		if err := s.compareHashAndPasswordFirebaseScrypt(hash, password); err != nil {
			return err
		}

		return nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrorPasswordInvalid
	}

	return nil
}

func (s password) compareHashAndPasswordFirebaseScrypt(hash, password string) error {
	input, err := ParseFirebaseScryptHash(hash)
	if err != nil {
		return fmt.Errorf("could not parse hash data: %w", err)
	}

	derivedKey, err := firebaseScrypt([]byte(password), input.salt, input.signerKey, input.saltSeparator, input.memCost, input.rounds)
	if err != nil {
		return fmt.Errorf("could not derive key: %w", err)
	}

	match := subtle.ConstantTimeCompare(derivedKey, input.rawHash) == 1
	if !match {
		return ErrorPasswordInvalid
	}

	return nil
}

func firebaseScrypt(
	password,
	salt,
	signerKey,
	saltSeparator []byte,
	memCost,
	rounds uint64,
) ([]byte, error) {

	// 1. scrypt step (Firebase uses N = 2^memCost)
	N := 1 << memCost

	fullSalt := append(salt, saltSeparator...)

	dk, err := scrypt.Key(
		password,
		fullSalt,
		N,
		int(rounds),
		1,
		FirebaseScryptKeyLen,
	)
	if err != nil {
		return nil, err
	}

	// 2. AES-CTR using dk as key
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, make([]byte, aes.BlockSize))

	// 3. Encrypt signerKey directly
	derived := make([]byte, len(signerKey))
	stream.XORKeyStream(derived, signerKey)

	return derived, nil
}

// Format and parsing implementation inspired by Supabase.
// See: https://github.com/supabase/auth/blob/v2.189.0/internal/crypto/password.go
func ParseFirebaseScryptHash(hash string) (*FirebaseScryptHashInput, error) {
	submatch := fbscryptHashRegexp.FindStringSubmatchIndex(hash)
	if submatch == nil {
		return nil, errors.New("crypto: incorrect scrypt hash format")
	}

	v := string(fbscryptHashRegexp.ExpandString(nil, "$v", hash, submatch))
	n := string(fbscryptHashRegexp.ExpandString(nil, "$n", hash, submatch))
	r := string(fbscryptHashRegexp.ExpandString(nil, "$r", hash, submatch))
	p := string(fbscryptHashRegexp.ExpandString(nil, "$p", hash, submatch))
	ss := string(fbscryptHashRegexp.ExpandString(nil, "$ss", hash, submatch))
	sk := string(fbscryptHashRegexp.ExpandString(nil, "$sk", hash, submatch))
	saltB64 := string(fbscryptHashRegexp.ExpandString(nil, "$salt", hash, submatch))
	hashB64 := string(fbscryptHashRegexp.ExpandString(nil, "$hash", hash, submatch))

	if v != "1" {
		return nil, fmt.Errorf("crypto: unsupported version %q", v)
	}

	memory, err := strconv.ParseUint(n, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid n parameter %q: %w", n, err)
	}
	if memory == 0 {
		return nil, fmt.Errorf("crypto: invalid n=0")
	}

	rounds, err := strconv.ParseUint(r, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid r parameter %q: %w", r, err)
	}
	if rounds == 0 {
		return nil, fmt.Errorf("crypto: invalid r=0")
	}

	parallelism, err := strconv.ParseUint(p, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid p parameter %q: %w", p, err)
	}
	if parallelism == 0 {
		return nil, fmt.Errorf("crypto: invalid p=0")
	}

	rawHash, err := base64.StdEncoding.DecodeString(hashB64)
	if err != nil {
		return nil, fmt.Errorf("invalid hash base64: %w", err)
	}

	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, fmt.Errorf("invalid salt base64: %w", err)
	}

	signerKey, err := base64.StdEncoding.DecodeString(sk)
	if err != nil {
		return nil, fmt.Errorf("invalid signer key: %w", err)
	}

	saltSeparator, err := base64.StdEncoding.DecodeString(ss)
	if err != nil {
		return nil, fmt.Errorf("invalid salt separator: %w", err)
	}

	input := &FirebaseScryptHashInput{
		v:             v,
		memCost:       memory,
		rounds:        rounds,
		parallelism:   parallelism,
		salt:          salt,
		rawHash:       rawHash,
		saltSeparator: saltSeparator,
		signerKey:     signerKey,
	}

	return input, nil
}

func (s password) RecoverPassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error {
	passwordPersister := s.persister.GetPasswordCredentialPersisterWithConnection(tx)

	passwordCredentialModel, err := passwordPersister.GetByUserID(userId)
	if err != nil {
		return fmt.Errorf("failed to get password credential by user id: %w", err)
	}

	if passwordCredentialModel == nil {
		err = s.CreatePassword(tx, userId, newPassword)
	} else {
		err = s.UpdatePassword(tx, passwordCredentialModel, newPassword)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s password) CreatePassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel := models.NewPasswordCredential(userId, string(hashedPassword))

	err = s.persister.GetPasswordCredentialPersisterWithConnection(tx).Create(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

func (s password) UpdatePassword(tx *pop.Connection, passwordCredentialModel *models.PasswordCredential, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel.Password = string(hashedPassword)
	passwordCredentialModel.UpdatedAt = time.Now().UTC()

	err = s.persister.GetPasswordCredentialPersisterWithConnection(tx).Update(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
