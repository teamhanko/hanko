package jwt

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	signatureKey := getSignatureJwk(t, key1)
	require.NotEmpty(t, signatureKey)
	verificationKeys := getVerificationJwks(t)
	require.NotEmpty(t, verificationKeys)

	jwtGenerator, err := NewGenerator(signatureKey, verificationKeys)
	assert.NoError(t, err)
	require.NotEmpty(t, jwtGenerator)
}

func TestGenerator_Sign(t *testing.T) {
	signatureKey := getSignatureJwk(t, key2)
	require.NotEmpty(t, signatureKey)
	verificationKeys := getVerificationJwks(t)
	require.NotEmpty(t, verificationKeys)

	jwtGenerator, err := NewGenerator(signatureKey, verificationKeys)
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
	signatureKey := getSignatureJwk(t, key1)
	require.NotEmpty(t, signatureKey)
	verificationKeys := getVerificationJwks(t)
	require.NotEmpty(t, verificationKeys)

	tests := []struct {
		Name    string
		JWT     string
		WantErr bool
	}{
		{
			Name:    "with signature key 1",
			JWT:     validJwt1,
			WantErr: false,
		},
		{
			Name:    "with signature key 2",
			JWT:     validJwt2,
			WantErr: false,
		},
		{
			Name:    "with unknown signature key",
			JWT:     invalidJwt,
			WantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			jwtGenerator, err := NewGenerator(signatureKey, verificationKeys)
			assert.NoError(t, err)
			require.NotEmpty(t, jwtGenerator)

			verifiedToken, err := jwtGenerator.Verify([]byte(test.JWT))
			if test.WantErr {
				assert.Error(t, err)
				assert.Empty(t, verifiedToken)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, verifiedToken)
				assert.Equal(t, subject, verifiedToken.Subject())
			}
		})
	}
}

func getSignatureJwk(t *testing.T, keyString string) jwk.Key {
	key, err := jwk.ParseKey([]byte(keyString))
	require.NoError(t, err)
	return key
}

func getVerificationJwks(t *testing.T) jwk.Set {
	key1 := getSignatureJwk(t, key1)
	key2 := getSignatureJwk(t, key2)
	set := jwk.NewSet()

	err := set.AddKey(key1)
	require.NoError(t, err)
	err = set.AddKey(key2)
	require.NoError(t, err)
	return set
}

var key1 = `{
  "d": "rPY8WpbbnGqd35cAGyDQ6EotozrQk7mjP8E6ztm0BLRoK0B8SYmGwxxpLZGDF6eeH5WSYl3a5RGvHZemMKZlMHryO_muMloRxu8iF9tfwjULcASu-1Ivz5wlXWe7PeFlOGwp3sbS1VwsUfkmDr6VCcRX7hEtmcn8QJyPs12b1gfBagOwZN0iNJYn3YUko5m-lGc4Ca-H3GXFPxtRt89z1h8B_7IRPDNOMZK3Wo1fl79MA_idIvt-Y5EAcUoi7aW8wlW0vx1h81r0taGUP7UOasz-2wkvMOJWU3Nn6GEx_S7fXfA7WFEr5KhoNLncdfNfVUKEuVPSIqkl6BClxovYPPb-ilZxHakIsrSFhcaQ_0OJpm8njmWCIIQNtZVgbV55OLzGaI6cDEmRFoEO_fvlG0rKxaGWDRJpJfZrt5VBhvxgNTDaJtOWyQRjC2d3y3rdIY0yxU113CYK5n1SeB3V1H--ElcOWlOeSh8pV_ZL1014TTWxBGpdgzvtHp-ooWs-qZZpnj_mb2bD-KdGvCczKm8QfFxQQgu8QswnE7LBnKQ3aM46M2swUlQnutjbMvc9HhNCVtwxTwdJTjCCkO6YClctxhvEjSksTQFY1Utr2jQsNWWLieRma1WUyH82_5BF4IF4jLQwZ6fZGtcJe0YJK2dlQQ4ITqHg8E_WsfsFSuE",
  "dp": "bv3Gj-UKF7QnUEXDLxbEPBPRt87tR8f-_9ELdCyDqDjZ_j1z4ua4vCZ1ldQ9bcaELG0NQVRFVDGD3EbaSAX2jWfxaKPIWuN7MV3ixvuf-bAdtPSRDzwaiflyG9icn9HWoz8TBL_VRwgTbswzXBTgqNnl2RpmCTxDcc4HeQexBSpwPNMV1pv5fHiP79GisGHBV5d5ycKaszcTkco-2Ls7KtBwL7pQBx-yCyHd9cQTuoL8ssOnkiB47KpZSohWZLbwjWmAK2-5_ufx5ByM5E60zAgbyVfO0xXxY5vjwSkbD4Vnr0tPxLScuAH3NUZ8rXr3t0aSgzKJxXKJueIiRtYVdw",
  "dq": "w_msGUgSOL4RLbU0FsieYoRAjLhYjfPJ4J4Pc1UuJOca-QREkE3PAti0KWNfZBm1WIdZphuZYUupt8-UhBIHk2Z8fhejSHuoBZXGj7tSvYNQtzTyzKrh9I8n1GxUop5KUIHRW70_I4-22OyUVhQsQtJewLG3-aG0qz3iYsscamNnNUZixAidhF4s66f1piWjhVTdqdp2NUF1fp41ebmHNajKiePoMQt5NQhzBZBxT9LmhB2e3nzNvTNNkrFocR7y__umk6fNfD7RUDcQGNl5L9RmvxgHL88KMi5ApOMiiJfAcEgtWpr2LvcBMgI7OrZnGodaYPE1ZG_sH_pqPYnAwQ",
  "e": "AQAB",
  "kid": "key1",
  "kty": "RSA",
  "alg": "RS256",
  "n": "ugADkMwTn1PIegp1VT9iOnX0L2-DhTL-6NO5LWx6GQMu_b87hODWL2v8u1fklUTNLEtmxubQY7nMsLUl0_zzHKUQLjk69vC5CMdq2yMYQ--sv2YViFQX4k9FMgNnveJunFQj4siA7VDIApxOYjSVMQ6HXQUCjclxMoQlLw6TvDd2h1dP6UXDvqgAp2iM8XaSiTWFtMgKkYreIB7aXwPgLrkffJ7wh7d9duLlygIIogGAucnxPLs8qn5Hpb9RwZs8kcXVgFWhZWlfnQaMXaVzjhN0uTcPD0JLpPQ9XLVtbsDwIRDYhTrZBX_58GQDlAJS-3M4LtkP3AXMnUhuWyK4vwCgoEV5VmPkWTF1DnwP2bjzRRzs0f9c6r6JYSEkZfIHksR8ThADnIUh1r89sRN5na5h9kC3g_9yLwUR3K64s1turZMZQsJSPxnXL-jMkaUKqgujMhH8A97aPnSRQ-SW7PHUGXE23tBEk_2rslsHOA22fo3Z09VyGwJD6BXwFwW9pj3tU4LNg0UyOk5QexSEOrVJS3dL3X9cXDgEqfhXlEZeFaA0882nxv2kwyUBsz5Rq2pHh7Pj4v0izWz-he4VyWg-8-HoNK3Oy4wGdhStTa2cU0CFNcu0VqTrVmYkpsEPNWmVapBlP-2pmiRLQLUYZVsjit9gIM4VwEPcna0wNcc",
  "p": "53kTdYIfPUEv_ml9vw6A0UPAdIf0plOvnCiHyzYQ7yY3Z37IfDy8h9esWvYH6xy7W65jECaBUcWn88UbMQ0lEaxsvh0XNunVmMtNX41bC-cRMH9kqcewvIHSRtjHbUmXCNpPETyNYRuUS7F0MbHh2y7SrG2krRSXaMABGamVZ0iWaKSUHXCCq-vuPnaEhoRDK3OI-aQQTr24I8LXmAECYyNjTC5NUiNAMSu7A3I7id31Pqgcht8v0gQSpWatJtg7RqNAA4S1QxnrjZRcpcq3IvrTWGB0WpgQb4z3yfKXxaz7XtLm8zSD2JfoRYTLM_qqbJwhvYzmSCFtDxxC1z8g9w",
  "q": "zbVxOg87-yWEOOd0k6jE2u25xpnUCfTpkLycApPXwtBSh4AQWpuTO4zUXbZUsQjg9mRZV1Ksnz1Vvb8BRx_Q60-G6lGVTANp1uP88HGA7hlIYxWCeZYDLfF0P0Iueb0a4D-kigX7UAwt5FzAbM4RtW-2VxcUIzP3lxxpsILw0arBVdkpv_wzjT4Mqz6KpQFmocdt9kqO6lVLn9dG90l-7CN9_El77aVpR13cqh-M7uDhgcN-SVwCZxmDvUcDC2Vj5JKEz4_SwBqPLMrH3Ba1ASeRvEJ7VVN4y_6zE6Wx4cQsk9bba9Yl_fzRo5yqW-JKCXksO_F2Sg4S_eDY-3ktsQ",
  "qi": "tkb9_3gwcIOmQqncjaT-A-vmKotEZs6U186u0CsQIKNvpcwMOWHjdRxd896tj9wE5F0JKxllwRqN2o0bb6QlmM56_C3Agh3MQ5CFiDC6Q7PEcSMbTihfRFUeNGP4gpiTaUg67HpdHWkYnsziYG6uTUh9onyntg-uZz3xMlOIEfURKCSYKamXgBQ_6EZyDG4xQv4Og_HUbCao4XThkiH4t70tkGi8hmagE0D7eiUY0PQul4x3Z9J9_xGdY3AxqvnLGnKbJw6m0g73fRcL3IixDfqFGuYIYvbdQCxgzURW3LGKyCsWx8A0CjVu9dU3XiNn86ZEVr4ZwxT9_XSbS1l3bA"
}`

var key2 = `{
  "d": "mF2NW0lRf03XFc7TIDZqscl-i2z5_-2jSc99SOACWn4NiM_Zh4VlVqe-hh1lJsE7bro8Kbe4L_h4ca5Ef_cHgd6oGjv5DPJidNDrcxtUpQoeJnTZuOcpAybseFjDLQCqhPca9zF8aeIEUq0F38YZqfu1d-94r10MtibgBBOvOpqkLlvjFVfxc_-fzpxIOnrrFC0Jtp1sjFbwTPeIV7HUmc8jkM1E8zXqyj9yQBHoFHUbEQHfK2Cdu_UXedzhKfpQovB1-y3-6Fw6VDbE1BChJ8BAp6zC2J7FdHOL47vqjMVE-io_NJcAm1P82hvFFfuitU1Nfi4rY8XWRuCwwhLp-aifQ5ugAUVjChzOCrka8eiN-BSlwmgiv-tql32DP7wYiY9ixGwdXRU7I4WH51O2L36Cqrq5Rrd0F0WfzRBmAiq4YPPjZp2w3p2pyvRz9jlEzE4BQXrdrcE6JY5Vb2htysSml3_aLXg--CgiO5D8R6IM8Oxh2bcE1cuZo67KFnO1_iOhyoNGXw9usclierniaHmdUVUwBujMzT3s9fE1zpEWccE1VpEzQ_8QUFxrlevmxMTLw5RolY71PruRuTXr135K8tMviOpuftWB05OtBhMKFWXTnq-qRMaSfp8nmHYpVk1qwDAe24f2s56ClkqiKkQnZsEgu5m_mEfcc_2Bi_U",
  "dp": "EBGwGpFQdAjGryNDr3heh1mAbLASvde9ARN8vmDRP9pMEDJOpnAatai-EYh38PpvUYKb6W_JLtwOshRa7U-6sjmZGaMLH5Ftbh4YF3nJ21RhyUczaP3CyOMaueeFPAP80e15RrZubZID5aIKF2H26wWvprzKGsh956uUI6AxtZUGdW4-N9_EzJkSBUqAlk7aO6KKZKYqoavUNZfYxst4dvy97-QI1s1dUFxPPTC6ltPk25Zv-n7nC4y9-Gb9GOO3oh3mfD2paQBX8blHwQU78ZZ2z2WCuXaGBUY41ybArFnFj8W8KhLPsocPoSkeuxSuOQnG7UW1qVWG3Nhxtupwnw",
  "dq": "q0s0U-b24CMeEnzKmyoYDRuIyUfmLzT7nnOaIxT_57tFt-7P9Nbp_HsMoAUoiV8eXP1byrxgqVLfJcsyDI408xSunj9wxCDSWwujg-aSf1gJ-mO9mtWTtttaDBGWeMgxIoA5DCpOzBQXXZ9CeaW95uzuWPrMr3xflYL__DSZ9wrrBs9aF2Ej3KspQYkg3kPUBC3SGvj5onXtbD9YV3mq-cpAyS2aD15UpIS4IwFuDT-iOe-23nc-wAc2rrC1RXNahrm9f7sXZrdEycOe8RQss3G0MBz8BEuQRvfegvqrRxjxHJ8P4CFnNDvu4vqsc-OWmaAzYqeonfHyVHWoRhzumQ",
  "e": "AQAB",
  "kid": "key2",
  "kty": "RSA",
  "alg": "RS256",
  "n": "sCi1KDOdUKttWUxSwA_yPE9cI4yb3d_Cio9ig2G2lk89ugOvzslqr6OtKdnjnKeYtKMMIDDbDTsa7kJVpWBHNxycIzoB1nUjXNYfYzqIfWZdXsbUlbTeTA-yUXyRW13CJQfxk59jUlYm1VLXxyZvScmBVmKlNLK7oDgHT72uYoGnem35w6Hb85gpuLnYR3E0iq7-H8yQ3oBKJ8TOyJwZSfzMujuz7njaZ0uW84phkb9H6We9-p_w9GSCeheY4Zu1Ercr77SbcI_hODxpyCjQUwgr7uGKae4P4moa1PkyBRWb2wSZYeqW81NkZlbu1TZfnKTaeQFSQMbUb4JNyk0Sz1lgfpgG6rdE1OTyIJmer6I9gRSgNmyxF9rGD-fpMk0ZPouLcbkx6KBmefWp-gv2l34-uW2czoXc7c5pmfPvWjxd1wLpES5SLca-zF5MNOhrtnqgZ-ComjajwgEztX_XZCy5vYK6qDnLOhognWOut0qJn1WMLgLvMgc9yfdZrNnZ803-zkiRIWR1uFkxw8uH2G-a2MqR6EJ8liN9HzSIUwwI5NTkkYyAndf69Rn-DuSPWpLnNNfgVz06fhdC6k6S-aGHrIrzLClLTnX9ni5j6ypdZcbA4iT_fcgBKD8BGJlV1QMzHvLCn8MKBoCNdgVBRS1vXTwwa61-7Cm3j9eWloE",
  "p": "ybVhcMu7a5bPmbcUYnXf6zYrjOU_yZjgae9R7ViVfBRkvMtU0G0a6EpR_wuxbCK6z-UzJLXxIC1h7NSnxhNRvEBMQeg1FVszab2Fl9k7PMNmtPqwbeQKVCKABmLn5vSErbA71D-aNUV0CxQKdvTJR19NvNjEHCyOM_MLxyNrBww2DnM1slTcE7Rr_OM2OI5q6lEA5FLtmtGof0qx7BrVabmJRhEASUPRZ8ir4mH05ZhUT5g5tVE5iZ_fTV8l5XbU_Xn5VqXo45kT3I9xuME8f2g2z9EIVZMGHoKg_ip0KeVSZtYS6tEQfdrwXpJoDq5TibuliXhSuGvht0yqaA4ZRw",
  "q": "35LaFXiFAIUerhgCPjlnXujNQ9pPXfgaCibp8qiBr6pqbgabYGT7l4d62IwwHwmFvcfhiFTlK7iqc9jiaGAaHDujrEuIWVTua4uWHUu2iCj_a8SnoKf0J1BgIdS85Z_VeQrjSbLtdDzKUARnzsVoNt9_y60RaEcL27Zyut5xfPnfQb2P_gSVrWa2LdDnOAESdMhgyq_GYqljWRU2V75DhS0P_FIpYifzc72AAkBTm1Ni7-DsoyN9rn8_wofhxHQ-NfuoRkd0WrQiL1Co9scSAWtevLpLqqFq76pgBvsJFYlkbPviDuTP4S5UYgTurbZZ69hW-PrQ-w60g_Iyg3u19w",
  "qi": "ttLq3IfagjTOoCLVuPKwmyHgt7FxzbpD0RevGDy25gwrSbmFma_zUFRmC361qcDk9qwSEC1V1qSXFGt5TnXp_ZYyJ6V0PxMa00kMYZqj1lAZ-PiVW8FlyV_tySHqG1gzBemj46gcuPXcHmzZ59CkkhPMzCfH_3QllM9H0Ts9oaW1rIXr2PUKW5GuttgXxDQ92DF-H-tHNx73yvSBcMA_LBn9sclGl6ARFr6sCVZrlq-tDsQXoWVRgLphA23L8O0PCaluhnh8cCBtlo8XlKfLJ5gIOrK7lzKtnCgZ-rP7lPBbo3vmvWCdbjQh5W4ht5aiHPczZlmZjPpSJ5n_7kgvxg"
}`

var subject = "c21ae0e1-39ad-494f-badd-2d54e072641e"
var validJwt1 = `eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleTEiLCJ0eXAiOiJKV1QifQ.eyJzdWIiOiJjMjFhZTBlMS0zOWFkLTQ5NGYtYmFkZC0yZDU0
ZTA3MjY0MWUifQ.YwN1KRYbAS2zhsI3mrOcH56ngMxHCKAEtGXrhAnDC0WxR577kB22upYFVWmHuAFJtd6gJRz7leHUIwy7Ed9uDgkHbOgGtxXXONbUVg-ho
l1qwJikJG_WFMFetlztY86U2lMHAwPgw8BazeEPmfSIq23KXuP1ifW33XJ_KnNloUCno44CXlsoUEpNKPIJULVLsf0UkuEzfhp0NwQJ0FcZ_Qh_g4QcJwLZ8
xmqnCoGSN7p9zBlxvMwietlPCAqI50S4wW5I5or9MpwHpo2ejrt2JLj1H5v6LtE8-FakGPE5Cz5_84tLbWhAPO_IqaN-xMC3O2LVrGik8AdltZCXnDBKToCf
u8LEUnX0wnuBp_LlooBpC5fo6mGN45q0MBEP3n6HXQpoMLZn50KVetG7YAuaFoYZLUd6I4bhOTUDMkCSK-thTL6_uMvFDhrMzxOAKjBfClq7rcpApCsIATEx
v2xMiK0Kj4rXXgL_0I5aHtiDyDviYz2LOnoZM36Pclei-ea399uZHT1caOfsZQCHCIMtc0I4wENCz34Du8y05CN5XNC_DrALOq8D0BVSESBdsmYAUMERx-u-
qBpal9KdTAdARKZvaAPT2Tc2ZlgXL42fkSVHjSMm51yJe6MB-KAa6a9FEbVYw7ZBnTf4_0aQ9StK8HwMjkx9E37NE3YjPbtOsg`
var validJwt2 = `eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleTIiLCJ0eXAiOiJKV1QifQ.eyJzdWIiOiJjMjFhZTBlMS0zOWFkLTQ5NGYtYmFkZC0yZDU0
ZTA3MjY0MWUifQ.iMhqgKf_ztfnGEbQX2qmt3o92BtfDJA6fvL3aKFWV7RC7XChH1-1OPj2AdsC73U3t8y8Ud4bcyengZffBu3L3wspD-MeJT8ti54yMWDqE
wAcA_9iOmk5X4SSBPphXGnuZIRWCvqdIzWnWk71ONu7fZ-VywN4MdkiqL1c5AevSxCK0AH7EFF6ZDg65QcFDfjSu6QI0HiRr7KU1uMK9BjdSWuvZlwogVI41
rE7LWqHdRSa56dQZoaHbbB-NEmGgC38eKo3BtIt6R_pCmMYNsKKxY5dHy6FM2WWmBwkYCPBxD6gowSXcPic2DTR__NlovbTCzVgEvIytHJhWlnX01d1Qd9Jk
5JOFb0bwkTpuxkwbZuxBBId7CDNBhVtnYIPhTY72yvC2lof7EFwoWVE1S_06JovgIBSlySCBN3kMNErswQQbbbJWge82WEMvOvC0GNFp8oGcHW-hygRA4u_T
GmWaQ1NUPAswqyYid32rsW2Pq6KjZcSh1c6rKaS7BTb3kMnRRiNDCn9Ibfe-2x1uWrkHur989PUa92ycTe89Tp_XSt9OXNqIE52auXmQYOOTyHrnmPTehzhv
cu1FSqGN1hfSK2RpQHjnxwrabekgqPbQA03KK0ZPTCG9lMQTd3zMsJutsfmXwfusVCwINTE0aqa4GSK4Y81ch4WAksznBH9S8A`
var invalidJwt = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjMjFhZTBlMS0zOWFkLTQ5NGYtYmFkZC0yZDU0ZTA3MjY0MWUifQ.c
WXSextdnNtzNll1MdrCG73qBM4px-pz-pCn1hCbG2g5aHLtKeKwxYAhus4i_NSVDMmuIULk9hmteUUAM3YByFtCjKZElWC9laEiYydERzatkJYi3-h1N05y
I8K2aav_3bPubThp_u0Mgwxiz10bx7Qon7BakvX27B29iETcWTAyMvrQTnQGC3Z89Z8plYeBUGkD4ftN8z3TSVUvdFvgJ8E1LnrricbL9mKigv2q9HMXqC_
23GmhSqRGdHp48JIyVf6PZoSD0qwC8mQNM3R_kMW9cCbTWu7CrdzNwsbB_NJoH_UXwteJMY19FeljeY3ELhWdy8tOzJwSz9G3oEFbtA`
