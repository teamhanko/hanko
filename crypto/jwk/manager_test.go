package jwk

import (
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/persistence/models"
	"testing"
)

type mockJwkPersister struct {
	jwks []models.Jwk
}

func (m mockJwkPersister) Get(i int) (*models.Jwk, error) {
	for _, v := range m.jwks {
		if v.ID == i {
			return &v, nil
		}
	}
	return nil, nil
}

func (m mockJwkPersister) GetAll() ([]models.Jwk, error) {
	return m.jwks, nil
}

func (m mockJwkPersister) GetLast() (*models.Jwk, error) {
	index := len(m.jwks)
	return &m.jwks[index], nil
}

func (m mockJwkPersister) Create(jwk models.Jwk) error {
	//increment id
	index := len(m.jwks)
	jwk.ID = index + 1

	m.jwks = append(m.jwks, jwk)
	return nil
}

func TestDefaultManager(t *testing.T) {
	keys := []string{"asfnoadnfoaegnq3094intoaegjnoadjgnoadng"}
	_, err := NewDefaultManager(keys, mockJwkPersister{jwks: []models.Jwk{}})
	require.NoError(t, err)
	//dm.GenerateKey()
}
