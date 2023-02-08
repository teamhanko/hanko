package session

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	manager := jwkManager{}
	cfg := config.Session{}
	sessionGenerator, err := NewManager(&manager, cfg)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)
}

func TestGenerator_Generate(t *testing.T) {
	manager := jwkManager{}
	cfg := config.Session{}
	sessionGenerator, err := NewManager(&manager, cfg)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	userId, err := uuid.NewV4()
	assert.NoError(t, err)

	session, err := sessionGenerator.GenerateJWT(userId)
	assert.NoError(t, err)
	require.NotEmpty(t, session)
}

func TestGenerator_Verify(t *testing.T) {
	sessionLifespan := "5m"
	manager := jwkManager{}
	cfg := config.Session{Lifespan: sessionLifespan}
	sessionGenerator, err := NewManager(&manager, cfg)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	userId, err := uuid.NewV4()
	assert.NoError(t, err)

	session, err := sessionGenerator.GenerateJWT(userId)
	assert.NoError(t, err)
	require.NotEmpty(t, session)

	token, err := sessionGenerator.Verify(session)
	assert.NoError(t, err)
	require.NotEmpty(t, token)
	assert.Equal(t, token.Subject(), userId.String())
	assert.False(t, time.Time{}.Equal(token.IssuedAt()))
	assert.False(t, time.Time{}.Equal(token.Expiration()))

	sessionDuration, _ := time.ParseDuration(sessionLifespan)
	assert.True(t, token.IssuedAt().Add(sessionDuration).Equal(token.Expiration()))
}

func TestGenerator_Verify_Error(t *testing.T) {
	manager := jwkManager{}
	cfg := config.Session{}
	sessionGenerator, err := NewManager(&manager, cfg)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	tests := []struct {
		Name    string
		Input   string
		WantErr error
	}{
		{
			Name:  "expired",
			Input: "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleTEiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2NTIxNzM4NTAsImlhdCI6MTY1MjE3Mzc5MCwic3ViIjoiYzU0YTZlZTUtZjNmMy00YzExLTlhMWMtYWUzOTgyYTQxMzYyIn0.AEzZ0M1_3HnOtqd8Dz-BliHkEUc4c5mu97eXhoErgG7qbVWisJP0qfz_KrwL9VYFOYuDAmfRZ3ABnaOg-S53wlRndfL-ulk68lY34otGZfXKhk2P3GJRH8Dq7hW83KnwkSPF5_iOaIIDfUwrWOaavvtLJFgg5RcehuwLkYEA5X17ek6cUNsqz7Vw-x2REReh_f31f5zneqKN9CeVnup5_ZgtMYpOXVvXAORs3b7y2oMwFdXs-hVal9ZVunNPo3iZmaTFMHUSNXX8MceOy_dUofxtd9JDzliiPrjNWDjU5Jx5paLBA5CUc4SctBURi2oJABbkeE1l4ug6-rTOYB04-UW8XAnPZONBTnv3AjtzvScvkpUj-OFKVQLGgcXZHUo1J7ftLaezpWrGTbhlC8TVvXdX1ms5w9D1uqEUZ94lhvVSW_lGGX2DGqMWaT6tOcSpDHFQ0NR5FD3MiNGV-z43AUOOSzilAKS2WaHDS7v43PeJ75xzAAS_7xOoc6L3Z9msdToQIauLYuCrivoOVcCqrEHugknpxO8M0xo6f1fHws8RocT3S7B76YJUIBeAj2F31wJne5xtbiRF5GWiV8uS3ZTXqrPp7y4U6Btf-h6mvEos_Q9o9w5hck-8lixUs5mObPDsT-W6PdEehRaSL7-13dy1GpB8wMP5fGlnRSff9y4",
		},
		{
			Name:  "wrong signature",
			Input: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAiLCJleHAiOjE2NDg2MjgzODAsImlhdCI6MTY0ODYyNDc4MCwiaXNzIjoiSG90M3RlOWlJbHVwbm5scDBld1E1RWJmIiwic3ViIjoiaU9ielRWdUc2bjl6Um1TY2ZMb0RFT1NiIn0.QZPEyEaGCJikNP2slVTGdsT3x8CuT4ynd5tdj-7c38Aa54277MPgGbapQ7JGrvwyjhAihzvvlqCxn2oX3zIFdu0HmSlxAXQ6Ah1K0KlQabneG8XNSed3sgp9xM0BYV1rB2SCuyXwE3U3zj5zFc4g4-v2Y1hpn7z4n3n9IlnnShK7NTUaaELlWPD8FQyp8mzZmJVSDoWbCMdywGHkX5ZWMUAwPfvC17kYZj6nqXC5ZJm3i2u-488cDeE5NxCFe-0ey14NtNtM9xTaPy5U8zvoqeCik1-ZNbxR_NJC4H25Cth2pm__e-W4KepGy7i-cLZ1T_DqNNk8HX9zX_Quj88FJw",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := sessionGenerator.Verify(test.Input)
			assert.Error(t, err)
		})
	}
}

func TestGenerator_DeleteCookie(t *testing.T) {
	manager := jwkManager{}
	cfg := config.Session{}
	sessionGenerator, err := NewManager(&manager, cfg)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	cookie, err := sessionGenerator.DeleteCookie()
	assert.NoError(t, err)
	assert.Equal(t, -1, cookie.MaxAge)
	assert.Equal(t, "hanko", cookie.Name)
}

var privateKey = `{
  "alg": "RS256",
  "d": "hsVXyJ1VjNFjiRqLE6bNZrAJDlnE33ptT4XpbPylfhlLfLB_OOB_YC5e4cBBoXlWaIJzYQ-qX0eSD2OdNg1JC0TgyQvwOqc9y6EKGyGu2asyHsJxLy8IiaqoqdqgiV0N_DsCYzt5Ew2nJq1P3XYqO5TJBpISixO47BEHaBgQeQBwSfmV3hmGYYTzJz6bwDNDhrBtg-2WiTfaq-3trorxo5Ww17-icr69Yad47Y4EIjKNL8SLPnWCt4NZTuT6Qs6QeUn-wOYPMaLh11DyZBNOuqiNWKjs_xPoi6C8jS1Jua0loTJXblDuMTDRL6-k82SByi7q8Yywr2TAdrotYbXF8kmaMmzW2gdQhkJs3xeNm3RyoIZOiU-7uzykSG8EkC0bx3mhIGVW_IOpzD2Xo2abbweR-PyX5z07qn9F1BHScdXViDmMNq2FU25D9K4FrRUqg8k5jpzkFrhcyPuw_hwB_BheNZgxulBbKy686qC6vTT41kZceD34PdBlMpzPctsK60GQWow8qs_OTQjD5ff_sTrNk4wzFpzo74ctcHOCZavW3gnZjhrMO9yHKGUBvgQJCiQ3C9nAkEP4pSOtk3nYgNLaWFftUYS_JKf36PcpM-YJYZEO33ayrcK19fp0aZbP11W1RpCs3jOaVWGwsS3xFE-4w_0xTbWoJACBgRENy_k",
  "dp": "qtwBo39K7eDKXoyXn1YUwk8hzaNwhDqfhWPMGHiPjS9W5PLdEpfaxkoMK38oiYkb0ohmEe_z54fmMTAD037lYAbQpW-Al8z1J0qfFEmSgmCVHL80u8Tvq6OtCJojJUDDMEBL-s67FGepXekjNCyS7S1zXJ_CFx619VQv5hadLga2p5TL8pYBjNBfS9FKFeZmaIF6tkz_fNEwud9kOXW1gOcpbgTmBZgDxlHbCiQcL44q7vdKjwHY1a6bi9cf9uvuJ7E_3ysWycTKaH3q0lveTe8I1ovZy2QbuvzEKuF9P_9B8rZiWYbPx6H9bPzd1TisEH9the3R86ILLkqZwGZhZQ",
  "dq": "zJ8N9S677F1s5YNYJ0LyzvL9bVcjaAwA-xUjDIxRhOMJWL_I-spBKfsOSwuFtr_KRSFj1ui25gIo8KJxsC1-1PBvjsM0OzjlvmaTqzsK1SFA6yt9Wh5VQ8BvqfH25g3JqHcAOqqYsymMFyq9c2ycaq9uYG-sxXiOYP3XCoZn_KsTnMZi0LLAL6A6BQoHSDDhnUHdPMrcZ8ePFjXowFKFlBWCOj0wWugpHc21TQFIeN9mfWAuyfEqyQP0G3FS60e3JW2B2NNZiui_o6lmSLnacLz51htpe23lgsUcJHkernow7-nOsWhvBdR90j69YUzowitL6WyJ_DEc1AehGYpOUw",
  "e": "AQAB",
  "kid": "key1",
  "kty": "RSA",
  "n": "2zNqGZKiogCjODBpzyRvwFlZ6hYaxJ4ZwYeFoN24eq_yHJB5OtEtgbUZ71lPkSqawLa-5qTtm1nBWY3ZAFVh9XC4fJsHWSIrwR7Mk9PLKWyAGFLyyGJy8srwdoxSUDbWa2CMeRsUaP_Syr_iytx1Kn9S9RRMrdC7PkMWaKx1KWQmIplrJx6qAiFlsTRvDFT0Ysfm0Vkti6xqTVYSc_bnObjLfiQ6UCKqF9fQDUJNFGLAeBAhkIRcBxp5G7PEiB5QOoRTrb4aqBIdxdjMWqjjfHTmm3EPqtIWsOjWRV1FsGyPkvolcBZXaNX-jf0oY1_7AryFujAFDslzGxg071yXRF9T_Brd1DW8paULQ2Vwkhc1d2c6Ioi-0D6255jlBKAVl-h3yedKWzYe5eyCijHZRs2jV1a3NX7ixzorcXH8GHB7PgM5lyZB5Rpf9-49MgW9Vo_b7nBCvEsN8uTc8jRyeG1zPTddAQ-tsMEmhsSa55EbQT6wk_nOu7xV-7eUAW8jwijJiDPOgDPmsOHtjoYx6BgcxCOYZ71s5g6qaKiCMecFpl7S1fxoIXcgjBNvv2Gzs6plRW74R6cVcohOfGVA7e0ULv1KOqJw6H-TjRmHBXQnw_K1biwYsL0SnE1Gu-iZC1_ktVuI8vf9k6m53HC_3_xrx0zqsad0fvIjpjRj2-E",
  "p": "_9QXKH2TREzUqChGiRrSrKeURTuufWRr8dePBurE5xbqd3Edc360J3jifwfxW9jGRUwehVEQFMAPToPQP3aVLwlroVg5CHFmt6BOChZJ1ZpYfNxvwIQDyxmcGtpGKHkMZMJj_C3XhYULz94ham8w9t3Ps5A2CTLs8erDtm_22zXw8nB7AeUMu0_QEJEtXrG12tMcsVUiG94QFx1udu2d_XortXQlEoFz4KMGQhYBQultOe1o7awgwBHhh9XdSzPifyYArk9qBQEKx-mPZsFFJ46e3IaF-pVfP15J5x4NOhTDRC_NX2ZlXIyiNw_X3cmpMBvgEuA9lY15dQtD0_iqGw",
  "q": "21kJj4Xvm2jX1-c8HIl4TAhPKI5470cPEx-8eViGO9KEsfc45T54a3shE3dP-YY6jQkpZritNzBnuSGaxSCJFhF63XZGYdh3p2GG73voO8dLqZNTFlitKaRA4UA_4byoimKdPaDR01Bhe9XCzIJCJfYqDGlTD2tIWcsytKwK0O9QkUqg-1ROlK02CMS4tBa8fzEXCYSnsB9iJUNOiLrHb6JdUUcOnCWvmnYFHIwhbH891Dhg9CMcCOwNWL1LGiCYilW-reM1pRHHB3H5b0_gwbg3DQ6dv4VmOCmzfNM0aTSvwkYQkfMQIF_SM8QWF6r9RSunMsoz_AKIjZ3yNj1xsw",
  "qi": "gQh-bEfYunCcUKXuaBNuyesAAI8F6tWgwMtXqr6X_Np_GvtDdjho2YP14Jtx2_kxDDZPSnP_h003kM6OdJdF469-s-AuRXeqX99yHMfWDYEkXxkp4WsmsKQgg5mQNsBr4d4zHyzsqc1ZKf2mL9zxb5dnVgQjVKrYGgsnBlfZeP-Cz_6c1CZ1YkoxiH52dNdQPfPJUTSUlIgRs2BgCbszQHOO6a1qwkQOjhhUX3-_KF6G4agT2NmZrb_O67GHzIoqXpWZykn93cJm5119BF9dAQbQx4vl0daMuPrh8UwMYx7GO3iNL5tl_wBc77Z4bZu8fn-XzHL4bb3mSjg5DqntKQ"
}`

type jwkManager struct{}

func (m *jwkManager) GenerateKey() (jwk.Key, error) {
	return nil, errors.New("not implemented")
}

func (m *jwkManager) GetPublicKeys() (jwk.Set, error) {
	key, err := getJwk()
	if err != nil {
		return nil, err
	}
	publicKey, err := jwk.PublicKeyOf(key)
	if err != nil {
		return nil, err
	}
	set := jwk.NewSet()
	err = set.AddKey(publicKey)
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (m *jwkManager) GetSigningKey() (jwk.Key, error) {
	return getJwk()
}

func getJwk() (jwk.Key, error) {
	return jwk.ParseKey([]byte(privateKey))
}
