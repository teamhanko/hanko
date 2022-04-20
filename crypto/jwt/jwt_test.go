package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	signatureKey := getJwk(t)
	require.NotEmpty(t, signatureKey)

	jwtGenerator, err := NewGenerator(&signatureKey)
	assert.NoError(t, err)
	require.NotEmpty(t, jwtGenerator)
}

func TestGenerator_Sign(t *testing.T) {
	signatureKey := getJwk(t)
	require.NotEmpty(t, signatureKey)

	jwtGenerator, err := NewGenerator(&signatureKey)
	assert.NoError(t, err)
	require.NotEmpty(t, jwtGenerator)

	token := jwt.New()
	err = token.Set(jwt.SubjectKey, subject)
	assert.NoError(t, err)

	signedTokenBytes, err := jwtGenerator.Sign(token)
	assert.NoError(t, err)
	require.NotEmpty(t, signedTokenBytes)
}

func TestGenerator_Verify(t *testing.T) {
	signatureKey := getJwk(t)
	require.NotEmpty(t, signatureKey)

	jwtGenerator, err := NewGenerator(&signatureKey)
	assert.NoError(t, err)
	require.NotEmpty(t, jwtGenerator)

	token := jwt.New()
	err = token.Set(jwt.SubjectKey, subject)
	assert.NoError(t, err)

	signedTokenBytes, err := jwtGenerator.Sign(token)
	assert.NoError(t, err)
	require.NotEmpty(t, signedTokenBytes)

	verifiedToken, err := jwtGenerator.Verify(signedTokenBytes)
	assert.NoError(t, err)
	require.NotEmpty(t, verifiedToken)
	assert.Equal(t, subject, verifiedToken.Subject())
}

func getJwk(t *testing.T) jwk.Key {
	pKey := getPrivateKey(t)
	key, err := jwk.FromRaw(pKey)
	require.NoError(t, err)
	return key
}

func getPrivateKey(t *testing.T) *rsa.PrivateKey {
	privateKeyBytes, err := base64.RawURLEncoding.DecodeString(privateKey)
	require.NoError(t, err)
	parsedPKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	require.NoError(t, err)

	pKey, ok := parsedPKey.(*rsa.PrivateKey)
	require.True(t, ok)

	return pKey
}

var privateKey = "MIIJQwIBADANBgkqhkiG9w0BAQEFAASCCS0wggkpAgEAAoICAQDXgVHvxGqPUlzVEc9KPufvOSVlmZM9lhEnWMt0XOq5m0wZOvqyxAqmXBOQUzK83MOfV3swwPJM50e0gaG49okNJDUvhYND_JRiA3FcFj85ZsjL9GqwOrc4KpRQVxTcm9w8HdbVmFF_OKYWUt--f9DKj3u21Y7NUMFNF1FzpXDXCGuLxqVknZL_Z4aovpTpHCuTzuASfG-XkvlnTmB5RLCUFagGwFmX1VbS0GNoh-vlgXDaLNjklyG_CQsdVmrr8O6N17W-CdKZJEyXULktd_IRSztPL1U1x1lAAGFpynDNQ1mkclHjN0IZpcjylx5i8rqt3-eJU58yq2HxfkdhozKnwtXWZQ1F1GGlHKAI_6hG1N1_8pIPuiSYy9LNVk_PoUrMxQ9LQbvPeq7yj2Cbso3zbb_h9v_AdBWIaWeX-1Hs3kfqbZhHzpR8WzjDPolNoeu2RXJNIMCyQn3kvBFqiiOSbdbyLIjVOmOUjBtFA217sNaHAcjW9CdsTnClQoktRTaQLNXgeWmD2J4KO2HnqcYLIb57-tea3i6RELlfQUlTFLptK_wXT5NCAbVCnOaq8S4UmZDaGNshUUvA-D7wRFZqTk6JgDMGtC_pU2GcKCO3VDwyGv6zEp83iSRTij_eesEvWVcEEWNz161l_VdXO7fpMPh_3PSyDdvq3hkh46ej8QIDAQABAoICAGMB4b_zEDXKVCX7qa1lmy73pSu5U8EemcDm9Yn_SkN9ioeo5haNJItrj_1li9Di5-jjyxAKBQe51eKjD8anVS25bcnoX_czKoShKkpxWhioFSZGo2FViGmAfmUurMHxxUvFNbcp5H87amqlJnAhzq3RH7hPAu1m5Xfid6RW5LGWB7rOx5ujHS7DxETwUf-K1qZwi9dSXf5YIscIZiAwo6NVE74OTtsHw3zVCmay03i8cDl8EyVqHbHjmLygwDynkyGNcczePGfpGlsGVh0Cly7EznnBuDcd3-4cfqSYwhw7jgqUDvUBpRedZ-Wz8dzpwUQysvAPf_tKa5QEPQ0pahJ6tJhhH-f80S7jqwXK08GhFBlpTdOE4yW8-6jKXV7SdgdACztVTLBnVmcghXrP_BuVoOhJppzRK8lG8qz652ofYMgOzOSxydP0Db_qmcF80NskXSkRFRVzt5i9iC4cqIxfVyks9J0QPt1B9Os2_f7nHslkSqS0lD9v_IsZfMJffHR8dM_WRIYbOYrkcvQz_z-5jSr18sqwFVkb9WxlrdpuHrdrpVV_tv2777DQI1GRtlOZpxOo3OXmjiybJ7Np51zaGnIw_mN_5s4x1EHz5jDGR-H_QW9GbQ9GM40RVs4pj5rHcNabby7ZzuKe0mpEuY7CU3BSjr5osBjl5mU9lwMlAoIBAQD4QjJkPL5pPpqfsnqCPFiAcikpjwabTgF3_xzsX9zj_gFDON4ZCdh2H0B_k7AGA3WModCSF2fCMOyzbtOI4FfhPyZZ7ay3R-BICltszwT5X5BEpjxQjniOICh4craZNApJf4dZg2MfXWBOdzodCaRrFiwCQRTWUh6sGr0H_8efDMQmdCeWISPcKyoO5Z_nUGnLJbEDpneOIAW-Wo7EY3OPac3FM0v0AE-5VOiqQ4qY-i0RkkSLwrCziinBslf8uHEATm_yNQjxTCPUatz1R97lmR-a6M9JcNO2338iTt0Ng6wSFaGza2JOVshNsZ0tGOAywdwMRtIbLym16-LJf7GzAoIBAQDeOaiYzUEsdsRaRdM1OzDsGuszTIG4tbLH3Wacx9CpwQ9OFnAltevLW8jBYJaRsOJHR1KPhc7FJ4bzO3Pfe_APULfnplMr6_bxB4y1Vz8jqO34XDIFkJKMWm6837ld-U8xjLtQ4E6RCepUoUCX-F72QZaVJcAq5uCujcJKXpR4TXTZo6xUPA0-FQH6w7ZXwADhxAX9Ebl5s5MxQeConYiO2noFjXIzYPlH5YkldIbmleR0GfwiNN_DyXiXzlbh4Pd21cfZ2JOJrsZiINPKTZOtbAQ_wshMdMALjSHWtCzCs0b4WyNCdyTAh6mz2xQuinavGuQTnkxxpV6SN3lGp9nLAoIBABfzsw7uuWRIEP0FaEJ2dgd2fDgxP27ueL_OEklP-mzYzeBhdTQvOf4zh7KHWj1KSiYWWpwtu-oFdGDfeXNESdZGlHmqr7ZDLgVlUmrOEmnI6Y9mBn2zMThtK9prHujrF2796d4eCgs1pBwN7sJscruOORLCmrMO2zy5m7FQ4T6cKbSYElWuvtn4JCepyeK0ZHCgI1L51aEVv9gcvpd-DOEyURMMnvBcs1RrN8NtnsqhoIWIeiqNzySTWPICNfEBDo38A1r3-PPm57IP2V-k3oGCY4U7nvwz8Yk8SPTTbQpnwMtB4QcBfkuWnd65GzQFqWPcRlG853qN81VE--1673cCggEBAMJxsRQChQRi52wVrLjnEeeFpkc8qkT0t3oqP57vN6VRSBMLjxVwGOHXbdHGsfjIzTWRMqxiaIoaC_rICpuB1ouQFVqcLipATdKYyIXj0VtidNbb1OkJlzE3761UFN4lRyYT_dLGcfh2tJNYhSx0JqNSwG_AmGTxn6ccYuSv3TlmjNfiXudVpECuIQ1KMkKVvi_NVXAaEjBq8GApRGpFbTeR8zLokQRj1bsTHO2pCGC6xyrPkc5cdW7a2qn54gvCzMUuSbBT0MSoKO2zy504Q_96hD1GMfy0K1XwJ6u1-3RhabfmBvQhTAcqrVKyXvZaMX8GCIsh98F48Ub_Qx6PwAECggEBAOhT9LA4dX_jGxQPBqmJDqa0gEX-RNen4JJPMwgVWpKNKY3uN2EJNdEScQtLbm_mGCHRl712XsfD9kJyygIE9P5X3S6O6CbdtuSKX1md7Vx9ed4tBAghrXBVtv-ZxOJ6Kn1oxRHnNyPbeZlEaAK70IpCJmpzDRINSnSx2WFSdCmh9TmB-jNbRPh32pZUa7_pzmzY81Qg9txLWBRNChI6dlaQYkSCi9aGA1rvkCt8zjvxu7hnyGQmq3Af5mhqhVGcg-QO4XKvT3cbhJQLshZqgTcVsoh03cCTWKRtwNhRZmtOwSULpG6nYuIG3harzh6aWG6NVy6qnJbxSi7tV13yot8"
var subject = "c21ae0e1-39ad-494f-badd-2d54e072641e"
