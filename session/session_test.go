package session

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	manager := jwkManager{}
	sessionGenerator, err := NewManager(&manager)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)
}

func TestGenerator_Generate(t *testing.T) {
	manager := jwkManager{}
	sessionGenerator, err := NewManager(&manager)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	userId, err := uuid.NewV4()
	assert.NoError(t, err)

	session, err := sessionGenerator.Generate(userId)
	assert.NoError(t, err)
	require.NotEmpty(t, session)
}

func TestGenerator_Verify(t *testing.T) {
	manager := jwkManager{}
	sessionGenerator, err := NewManager(&manager)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	userId, err := uuid.NewV4()
	assert.NoError(t, err)

	session, err := sessionGenerator.Generate(userId)
	assert.NoError(t, err)
	require.NotEmpty(t, session)

	token, err := sessionGenerator.Verify(session)
	assert.NoError(t, err)
	require.NotEmpty(t, token)
	assert.Equal(t, token.Subject(), userId.String())
	assert.False(t, time.Time{}.Equal(token.IssuedAt()))
	assert.False(t, time.Time{}.Equal(token.Expiration()))
}

func TestGenerator_Verify_Error(t *testing.T) {
	manager := jwkManager{}
	sessionGenerator, err := NewManager(&manager)
	assert.NoError(t, err)
	require.NotEmpty(t, sessionGenerator)

	tests := []struct {
		Name    string
		Input   string
		WantErr error
	}{
		{
			Name:  "expired",
			Input: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NTA0NDM0OTgsImlhdCI6MTY1MDQzOTg5OCwic3ViIjoiNzM0ZjdlYTQtNmM1MC00OTMxLTkwNDAtOGFhNTBjMWRmZTBhIn0.tMa3NGDOK8D_mKYP0ahaNOiIn9hrjvpFg191WYoCK9_ET3MUXp07qBIyhM0sE5sYv2Th0VJ6KLlyfpCsuJxDvIaUXqYd0z4uxokC64nxF6XA7TrNFI8kKgRgAlvFqCNW4Z0qanipmHsae1sKEj423jGU_4EZVWtF-_7v8sXIpNmIYpn4pxkcj-qZiEaNWO6QZhvSyUIlU9ZW1_6o8UQMM_A6D8FMR7h_rjQdE7pV6JkCacMns5ge6umcWmCA0HrHL3_3-5-GRUdLoUwAraJLTXZBUuwCES4i99LOB0J6oo5LTsCs1LA8AFCdVU47q_AFvhaB7oR8a1HUffUvlraverx9dFSu0g5WqwqlLYanh_O1-F67ClG-v2KyQ81YCLa9tQDiUB0OE2pH6-3-Hcd-A3jTHjVbL3DW9Qkb9y1GpcD91oU64tThYM-8as7xvzvb-_oFvCGFp_BI10eVuSzZGV5DV4OMcop73SV5MJWZmz7HVsufr5WoE3HpSUBTsKxGP6ZJfJNT_muXxDvhLg7FBHeQ6fBwm2P-ONDxXuFclxJ7vGbIdF-YiGVlSWmq2Zp0LVF3WkUQv8B8cRadMgpuli6vIRvrd8aS7IDnVDec2i4Tj8bOov0fRsgK12Y56CUUWIm9J8s2hhzwD9FSrYrY92tRGzbMqFbhqZ8kZceRfa4",
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

var privateKey = "MIIJQwIBADANBgkqhkiG9w0BAQEFAASCCS0wggkpAgEAAoICAQDhZLLiFJXZc37mRZ_qeXD11Y5ZbFaqZotXV0AqVCz9gVBGo2rBeVRpap_z4ADPJVJp_1dYLg52lNKRmNsqGYSCqUZyiRNaH0Pf-RmFjzwci4BDaECvD8v-EKuXZCScUTtAJaMmT0xLwsm7bNPGHqN6JICe6gh53LA9Q0bCwbYttaiKlOB8yXd3d5_-ZMRFyjXltKt7amadxpXtAAEtcgyy-y2j4Sy8xysSjr4O7XRTZmi7-zCJAlg24fTGZUITAX4wpah6tG4ATzwTwmmUbmZN1h9IGGtXSTNs1Cp81F6Jk3RdgqAq6GCY9u3wSGJ8kfygD38LjwxN3UsyLNLufNQ1Z_U9kdhQtP4TWChons7qkBKTIJVqGZ9whUnwelD7Khq4LbiSytEsMCiCib60-ybz8YClT-NTVH9j2xYyhOyh4uINpcOO2kP9gU6EYI5otPSqvMueGwBqWp4I15qdGuzHuLsFxOoOhSu9QL9diLFPOgsO4q0nzuZ6bSonnLQsR_8O-Yxak-zcWUMVjTCEz6Vs94SoS9OAgxGQhzA0SQ55k_6Z2eiU5gJLWSPXae-s89UIESf8sPVu7anQ_WZPLVs8hb09BBZvoU5oJs7Pna9p9As-pA7ENgwJiK3hzZ4ZT9Tvvb5GLl_1gPbVtJTCZuQF2aBTM-mD_SwWZe_cP33MHwIDAQABAoICAQCANYanoWwH0HHLzKkFeGTwAbVCWqUFsuTqHsBqE42v-gHO3KAaQ8jnWfZ4g-AR9Lnnf46Qo0oo28jXdyqbzP4aUO24sw5mAkjau1hwJ6Ta2-Nu9Htu2T6BW7wvlpBYtsBMYdxnK05L_hZAXcws8zqsfN0JCDkgEI_TmVRD7mqRn7aqdbsoYHVraImC7JDU3gxAiL_OqRyL_O1Fbe49ipV8rfItOSX4kBaJLNchqKK12hgTbfQSy1mghnF09R5br0q3o1Ot0LqNxIR4_OqPuyjId9c9bF6KvSHaculkLm1ENrNHiclP_vULrdJ1Dseu8l_QMGBlE8687_cZKHQnoqwWT6_0zPuOpmZZiig7lZTGY1vlt-Jt6NCh5-YjptjIfwMSFouDa7FEzqUmKaG-8B3ijL7Bo7FhFhJbN2PYfVUWkVz2_5tx6CcVjW_v8hcIQLC_MJmKUucAeLcQeJrfirCLtLQeTQZOS7U16r1JYCx22O5LECEv5J2hXq7ajxs7DPv5-KfHD-ZcamUSjRT2U1KyA3mZxGbWmWu_nV_tRH3briWidBrHI-l8ZERWoW6eV_kEd4uG5taZmUPEtj-1v32_8hhJAomHseCxYM_m0q-joAdr1TXrT0O_zzKoYgXJpmfJet_HHQS5thFoFyky8i6gCMVf6zdycTygSheUDFM3MQKCAQEA8FNQcSJuNj9Mt86tn3mfX-8qcOy7l0LpDtKgT5_HgPk6MbUXgVWWZbs3kVboqy57NznEXfb9T1ps8hvDTpN17ErS4GQPn3Xufc7pMiMc11uC0H-aKODiHGB_XFlgtGG9n86sjZ62a8VsVBj4KUIWdVrq76S8kK_xMaU9xvz2Zxv378A-OfWv5es3SRyHoog_v9CAbPo-ZA7qenLT47BPVEJOYTU8-_QKC-vGMkIPVHM_jgMFt6V9fQFhjRy5Htbd7RyEv7B3yz37Ql_NHpKBuxX0sQ7Q-IH7-9dw8VBi-n8rp2dveLLgiEqrutENR3Mod7wL6inbN4yKgQl6OFwhOQKCAQEA8BgQreq-Xgz_9qbBCO5T--JiK7Z1ZMDk70B36fIilWCGXzEIE18xZM6tUq50_2gUHqiYbMo1NnKIoCbfW7jhNFmg_cNOqwcZX2JewdfNN4fBZPDOfu58CsuDiWtboqMHzxNDuIscNDZdj28fIe74kJKHhRofSzkAYAtA6Q_UbOooFcS5QQVruBt-I4D0MuxlpMizV4_ZPNF1hpYlpQ2EJYkFSMXx3umzAKU6XC6jE_E4kr8fPYGhd2Db-wAu8DrFexAt-JvfoHFhdkFfjlIJOKtfOCIKu0o1aguoPUXyoB88HQhsRv4hNcTEkJJaSB4tFGSs6ur4_vvkYQ8FjH1QFwKCAQEAgXxyWDKz7TiX7mVGeSl_rKHhXSzAOlTL27eytpQhWyVtrIClJINn4HJKE14fSLRnoS7X1cURYOMY1i4NQlYDcIg0LMDdBg71rAWC8genL4XX6t0Fw8a_LYj0tl5V03riP6uMn1WHdnPN1VYKx7ga_6o38VzyWIbjztr4eTGs1YtlQGF1Zacx2hCtHhBoKDN_HauKtqzyVtkOj2E1N3W0mHKNZqTXse0gSKIFjOi498iM0shgGT3qaiMHW4_BUpN0yZ_XCq1bLj-8FFwn2bQYgCPpTkjsYSkwCtZevTaRzeQdMjpx_jdq8SRCeQrQO9IZWMISLV3WBo0Lx1DC8ID7SQKCAQB8ySUyH0WeAEew3G6Lw8LmsXywl35gRVk3eFxavTx4QtjT9Nnrp5g2eqzewkmQzXlXjeza7iXGDLUx98IzG94ApWzlN3NVtLTdPHVfblf8upQrcHUCx9S0j53n-GKCHxXZ7HtQGQ0pne_2spyNuHR8P4wsA62sHQ4y3OZ5u35-tRzsY3idcMHRyIhAz68cIH2brax4oA0abQsWTkd2h5XdJGAYuLjDUNd2SGoSqzKbFM6AhtEn2I4hS7hJtoiu1vz4vyoFgo4yB3vOSJ_vine8emVc-WR2f1VProtcfVRjIJjPxibwpvh_x6saMNa3kOeBJ-ovlryfWjASugn6QM81AoIBACH3YhSfwnpC6FIQ77D4YmruMQoECnqA33dZarRQGlfUuNo0JKk-7yWNLAyE1HKRrkNWlmbIzU6Fi-BbgnpuzpnX2k-jRNBLUm3m7xzo_NBck4d-qPzQGn31vJ1bmZUd-FVPOzl3Pq8gv4h0kbJ-st7z90bKFjhD1_haKXbe-lyX-4d5a56ce5jsOnYCklx0Q1VTThuTCvAnD6XkvtVWAJ682-e-cWryQHfmXDpVLOler4D_PTzYDcH-Rzj1sRig-IvJS2cQNCqkh7Uf1y7xtJg7OPBlCvkprghAOCqgltKpr5Tot_pQEpZp9jJtPNBkpyDDZUbRf0MNQKcq_fZZw5s"

type jwkManager struct{}

func (m *jwkManager) GenerateKey() (jwk.Key, error) {
	return nil, errors.New("not implemented")
}


func (m *jwkManager) GetPublicKeys() ([]jwk.Key, error) {
	key, err := getJwk()
	if err != nil {
		return nil, err
	}
	publicKey, err := jwk.PublicKeyOf(key)
	if err != nil {
		return nil, err
	}
	return []jwk.Key{publicKey}, nil
}

func (m *jwkManager) GetSigningKey() (jwk.Key, error) {
	return getJwk()
}

func getJwk() (jwk.Key, error) {
	pKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}
	key, err := jwk.FromRaw(pKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func getPrivateKey() (*rsa.PrivateKey, error) {
	privateKeyBytes, err := base64.RawURLEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}

	parsedPKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	pKey, ok := parsedPKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("parsed key is not type of rsa.PrivateKey")
	}

	return pKey, nil
}
