package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/test"
)

func TestMetadataAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(metadataAdminSuite))
}

type metadataAdminSuite struct {
	test.Suite
}

func (s *metadataAdminSuite) TestMetadataAdminHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		userId             string
		expectedStatusCode int
		expectedMetadata   *admin.Metadata
	}{
		{
			name:               "should return metadata for user with metadata",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Public: json.RawMessage(`{
					"existing_public_str": "data",
					"existing_public_num": 1,
					"existing_public_bool": true
				}`),
				Private: json.RawMessage(`{
					"existing_private_str": "data",
					"existing_private_arr": [
						"existing_private_arr_0",
						"existing_private_arr_1"
					]
				}`),
				Unsafe: json.RawMessage(`{
					"existing_unsafe_str": "data",
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value"
					}
				}`),
			},
		},
		{
			name:               "should return no content for user without metadata",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedStatusCode: http.StatusNoContent,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on empty userID",
			userId:             "",
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on non existing user",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			expectedStatusCode: http.StatusNotFound,
			expectedMetadata:   nil,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			err := s.LoadFixtures("../test/fixtures/metadata")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/users/%s/metadata", currentTest.userId),
				nil,
			)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code)

			if currentTest.expectedStatusCode == http.StatusOK {
				var metadataResponse *admin.Metadata
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &metadataResponse))

				if currentTest.expectedMetadata == nil {
					s.Nil(metadataResponse)
				} else {
					s.NotNil(metadataResponse)

					s.Require().Equal(
						gjson.GetBytes(currentTest.expectedMetadata.Public, `@pretty:{"sortKeys":true}`).String(),
						gjson.GetBytes(metadataResponse.Public, `@pretty:{"sortKeys":true}`).String(),
					)
					s.Require().Equal(
						gjson.GetBytes(currentTest.expectedMetadata.Private, `@pretty:{"sortKeys":true}`).String(),
						gjson.GetBytes(metadataResponse.Private, `@pretty:{"sortKeys":true}`).String(),
					)
					s.Require().Equal(
						gjson.GetBytes(currentTest.expectedMetadata.Unsafe, `@pretty:{"sortKeys":true}`).String(),
						gjson.GetBytes(metadataResponse.Unsafe, `@pretty:{"sortKeys":true}`).String(),
					)
				}
			}
		})
	}
}
func (s *metadataAdminSuite) TestMetadataAdminHandler_Patch_Errors() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		userId             string
		patchMetadata      json.RawMessage
		expectedStatusCode int
		expectedMetadata   *admin.Metadata
	}{
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			patchMetadata:      json.RawMessage(`{"public_metadata":{"key":"value"}}`),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on empty userID",
			userId:             "",
			patchMetadata:      json.RawMessage(`{"public_metadata":{"key":"value"}}`),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on non existing user",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			patchMetadata:      json.RawMessage(`{"public_metadata":{"key":"value"}}`),
			expectedStatusCode: http.StatusNotFound,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on invalid JSON",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata:      json.RawMessage(`{"public_metadata":"key":"value"`),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on top level empty string as metadata patch",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata:      json.RawMessage(`""`),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on top level string as metadata patch",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata:      json.RawMessage(`"invalid"`),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on top level array as metadata patch",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata:      json.RawMessage(`["in", "valid"]`),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
		{
			name:               "should fail on exceeding metadata size limit (3000)",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata:      json.RawMessage(metadataExceedingLimit),
			expectedStatusCode: http.StatusBadRequest,
			expectedMetadata:   nil,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			err := s.LoadFixtures("../test/fixtures/metadata")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/users/%s/metadata", currentTest.userId),
				bytes.NewReader(currentTest.patchMetadata),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code)
		})
	}
}

func (s *metadataAdminSuite) TestMetadataAdminHandler_Patch() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		userId             string
		patchMetadata      json.RawMessage
		expectedStatusCode int
		expectedMetadata   *admin.Metadata
	}{
		{
			name:               "should do nothing on empty patch object",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata:      json.RawMessage(`{}`),
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Public: json.RawMessage(`{
					"existing_public_str": "data",
					"existing_public_num": 1,
					"existing_public_bool": true
				}`),
				Private: json.RawMessage(`{
					"existing_private_str": "data",
					"existing_private_arr": [
						"existing_private_arr_0",
						"existing_private_arr_1"
					]
				}`),
				Unsafe: json.RawMessage(`{
					"existing_unsafe_str": "data",
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value"
					}
				}`),
			},
		},
		{
			name:   "should do nothing on empty patch object for top level keys",
			userId: "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata: json.RawMessage(`{
				"public_metadata": {},
				"private_metadata": {},
				"unsafe_metadata": {}
			}`),
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Public: json.RawMessage(`{
					"existing_public_str": "data",
					"existing_public_num": 1,
					"existing_public_bool": true
				}`),
				Private: json.RawMessage(`{
					"existing_private_str": "data",
					"existing_private_arr": [
						"existing_private_arr_0",
						"existing_private_arr_1"
					]
				}`),
				Unsafe: json.RawMessage(`{
					"existing_unsafe_str": "data",
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value"
					}
				}`),
			},
		},
		{
			name:   "should ignore unknown top level keys",
			userId: "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata: json.RawMessage(`{
				"public_metadata": {
					"new_key":"new_value"
				},
				"unknown_metadata": {
					"key":"value"
				}
			}`),
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Public: json.RawMessage(`{
					"existing_public_str": "data",
					"existing_public_num": 1,
					"existing_public_bool": true,
					"new_key": "new_value"
				}`),
				Private: json.RawMessage(`{
					"existing_private_str": "data",
					"existing_private_arr": [
						"existing_private_arr_0",
						"existing_private_arr_1"
					]
				}`),
				Unsafe: json.RawMessage(`{
					"existing_unsafe_str": "data",
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value"
					}
				}`),
			},
		},
		{
			name:   "should merge new fields with existing metadata",
			userId: "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata: json.RawMessage(`{
				"public_metadata": {
					"existing_public_str": "data_updated",
					"new_key": "new_value"
				},
				"private_metadata": {
					"existing_private_arr": [
						"existing_private_arr_0_updated"
					],
					"new_key": "new_value"
				},
				"unsafe_metadata": {
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value_updated"
					},
					"new_key": "new_value"
				}
			}`),
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Public: json.RawMessage(`{
					"existing_public_str": "data_updated",
					"existing_public_num": 1,
					"existing_public_bool": true,
					"new_key": "new_value"
				}`),
				Private: json.RawMessage(`{
					"existing_private_str": "data",
					"existing_private_arr": [
						"existing_private_arr_0_updated"
					],
					"new_key": "new_value"
				}`),
				Unsafe: json.RawMessage(`{
					"existing_unsafe_str": "data",
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value_updated"
					},
					"new_key": "new_value"
				}`),
			},
		},
		{
			name:   "should clear specific metadata fields when sending null",
			userId: "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata: json.RawMessage(`{
				"public_metadata": null,
				"private_metadata": {"existing_private_str": null},
				"unsafe_metadata": {"existing_unsafe_str": null}
			}`),
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Private: json.RawMessage(`{
					"existing_private_arr": [
						"existing_private_arr_0",
						"existing_private_arr_1"
					]
				}`),
				Unsafe: json.RawMessage(`{
					"existing_unsafe_obj": {
						"existing_unsafe_obj_key": "existing_unsafe_obj_value"
					}
				}`),
			},
		},
		{
			name:   "should clear all metadata fields when sending null for all fields",
			userId: "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			patchMetadata: json.RawMessage(`{
				"public_metadata": null,
				"private_metadata": null,
				"unsafe_metadata": null
			}`),
			expectedStatusCode: http.StatusNoContent,
			expectedMetadata:   nil,
		},
		{
			name:   "should create metadata for user without metadata",
			userId: "38bf5a00-d7ea-40a5-a5de-48722c148925",
			patchMetadata: json.RawMessage(`{
				"public_metadata": {"public_key": "public_value"},
				"private_metadata": {"private_key": "private_value"},
				"unsafe_metadata": {"unsafe_key": "unsafe_value"}
			}`),
			expectedStatusCode: http.StatusOK,
			expectedMetadata: &admin.Metadata{
				Public:  json.RawMessage(`{"public_key":"public_value"}`),
				Private: json.RawMessage(`{"private_key":"private_value"}`),
				Unsafe:  json.RawMessage(`{"unsafe_key":"unsafe_value"}`),
			},
		},
		{
			name:               "should clear all metadata when sending null",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			patchMetadata:      json.RawMessage(`null`),
			expectedStatusCode: http.StatusNoContent,
			expectedMetadata:   nil,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			err := s.LoadFixtures("../test/fixtures/metadata")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/users/%s/metadata", currentTest.userId),
				bytes.NewReader(currentTest.patchMetadata),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code)

			if currentTest.expectedStatusCode == http.StatusOK ||
				currentTest.expectedStatusCode == http.StatusNoContent {

				// Verify the response contains correct updated metadata
				var metadataResponse *admin.Metadata
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &metadataResponse))
				if currentTest.expectedMetadata == nil {
					s.Nil(metadataResponse)
				} else {
					s.NotNil(metadataResponse)
					s.Equal(
						gjson.GetBytes(currentTest.expectedMetadata.Public, `@pretty:{"sortKeys":true}`).
							String(),
						gjson.GetBytes(metadataResponse.Public, `@pretty:{"sortKeys":true}`).
							String(),
					)
					s.Require().Equal(
						gjson.GetBytes(currentTest.expectedMetadata.Private, `@pretty:{"sortKeys":true}`).
							String(),
						gjson.GetBytes(metadataResponse.Private, `@pretty:{"sortKeys":true}`).
							String(),
					)
					s.Require().Equal(
						gjson.GetBytes(currentTest.expectedMetadata.Unsafe, `@pretty:{"sortKeys":true}`).
							String(),
						gjson.GetBytes(metadataResponse.Unsafe, `@pretty:{"sortKeys":true}`).
							String(),
					)

					// Also verify the metadata was actually updated in the database
					userID, err := uuid.FromString(currentTest.userId)
					s.Require().NoError(err)
					metadataModel, err := s.Storage.GetUserMetadataPersister().Get(userID)
					s.Require().NoError(err)
					if currentTest.expectedMetadata == nil {
						s.False(metadataModel.Public.Valid)
						s.False(metadataModel.Private.Valid)
						s.False(metadataModel.Unsafe.Valid)
					} else {
						if currentTest.expectedMetadata.Public != nil {
							s.True(metadataModel.Public.Valid)
							s.Equal(
								gjson.GetBytes(currentTest.expectedMetadata.Public, `@pretty:{"sortKeys":true}`).
									String(),
								gjson.GetBytes([]byte(metadataModel.Public.String), `@pretty:{"sortKeys":true}`).
									String(),
							)
						} else {
							s.False(metadataModel.Public.Valid)
						}
						if currentTest.expectedMetadata.Private != nil {
							s.True(metadataModel.Private.Valid)
							s.Equal(
								gjson.GetBytes(currentTest.expectedMetadata.Private, `@pretty:{"sortKeys":true}`).
									String(),
								gjson.GetBytes([]byte(metadataModel.Private.String), `@pretty:{"sortKeys":true}`).
									String(),
							)
						} else {
							s.False(metadataModel.Private.Valid)
						}
						if currentTest.expectedMetadata.Unsafe != nil {
							s.True(metadataModel.Unsafe.Valid)
							s.Equal(
								gjson.GetBytes(currentTest.expectedMetadata.Unsafe, `@pretty:{"sortKeys":true}`).
									String(),
								gjson.GetBytes([]byte(metadataModel.Unsafe.String), `@pretty:{"sortKeys":true}`).
									String(),
							)
						} else {
							s.False(metadataModel.Unsafe.Valid)
						}
					}
				}
			}
		})
	}
}

const metadataExceedingLimit = `{
  "public_metadata": {
    "users": [
      {
        "id": 1,
        "name": "Alice",
        "email": "alice@example.com",
        "active": true,
        "roles": [
          "admin",
          "editor"
        ],
        "settings": {
          "theme": "dark",
          "language": "en"
        },
        "preferences": {
          "newsletter": true,
          "notifications": {
            "email": true,
            "sms": false
          }
        },
        "activity": [
          {
            "timestamp": "2025-05-07T10:00:00Z",
            "action": "login"
          },
          {
            "timestamp": "2025-05-06T08:32:00Z",
            "action": "update_profile"
          }
        ]
      },
      {
        "id": 2,
        "name": "Bob",
        "email": "bob@example.com",
        "active": true,
        "roles": [
          "user"
        ],
        "settings": {
          "theme": "light",
          "language": "fr"
        },
        "preferences": {
          "newsletter": false,
          "notifications": {
            "email": true,
            "sms": true
          }
        },
        "activity": [
          {
            "timestamp": "2025-05-06T22:00:00Z",
            "action": "logout"
          },
          {
            "timestamp": "2025-05-05T14:12:00Z",
            "action": "purchase"
          }
        ]
      },
      {
        "id": 3,
        "name": "Carol",
        "email": "carol@example.com",
        "active": false,
        "roles": [
          "user"
        ],
        "settings": {
          "theme": "dark",
          "language": "es"
        },
        "preferences": {
          "newsletter": true,
          "notifications": {
            "email": false,
            "sms": false
          }
        },
        "activity": [
          {
            "timestamp": "2025-05-04T19:40:00Z",
            "action": "deactivate_account"
          }
        ]
      },
      {
        "id": 4,
        "name": "Dan",
        "email": "dan@example.com",
        "active": true,
        "roles": [
          "user",
          "moderator"
        ],
        "settings": {
          "theme": "light",
          "language": "de"
        },
        "preferences": {
          "newsletter": true,
          "notifications": {
            "email": true,
            "sms": false
          }
        },
        "activity": [
          {
            "timestamp": "2025-05-03T08:00:00Z",
            "action": "flag_post"
          }
        ]
      },
      {
        "id": 5,
        "name": "Eve",
        "email": "eve@example.com",
        "active": true,
        "roles": [
          "editor"
        ],
        "settings": {
          "theme": "dark",
          "language": "it"
        },
        "preferences": {
          "newsletter": false,
          "notifications": {
            "email": false,
            "sms": true
          }
        },
        "activity": [
          {
            "timestamp": "2025-05-01T13:20:00Z",
            "action": "edit_post"
          },
          {
            "timestamp": "2025-04-30T09:45:00Z",
            "action": "comment"
          }
        ]
      },
      {
        "id": 6,
        "name": "Frank",
        "email": "frank@example.com",
        "active": false,
        "roles": [
          "user"
        ],
        "settings": {
          "theme": "light",
          "language": "en"
        },
        "preferences": {
          "newsletter": true,
          "notifications": {
            "email": true,
            "sms": false
          }
        },
        "activity": [
          {
            "timestamp": "2025-04-29T10:00:00Z",
            "action": "unsubscribe"
          }
        ]
      },
      {
        "id": 7,
        "name": "Grace",
        "email": "grace@example.com",
        "active": true,
        "roles": [
          "admin"
        ],
        "settings": {
          "theme": "dark",
          "language": "en"
        },
        "preferences": {
          "newsletter": false,
          "notifications": {
            "email": false,
            "sms": false
          }
        },
        "activity": [
          {
            "timestamp": "2025-04-28T11:30:00Z",
            "action": "ban_user"
          }
        ]
      },
      {
        "id": 8,
        "name": "Hank",
        "email": "hank@example.com",
        "active": true,
        "roles": [
          "user"
        ],
        "settings": {
          "theme": "light",
          "language": "pt"
        },
        "preferences": {
          "newsletter": true,
          "notifications": {
            "email": true,
            "sms": true
          }
        },
        "activity": [
          {
            "timestamp": "2025-04-27T07:00:00Z",
            "action": "like_post"
          }
        ]
      },
      {
        "id": 9,
        "name": "Ivy",
        "email": "ivy@example.com",
        "active": false,
        "roles": [
          "editor"
        ],
        "settings": {
          "theme": "dark",
          "language": "ja"
        },
        "preferences": {
          "newsletter": false,
          "notifications": {
            "email": true,
            "sms": false
          }
        },
        "activity": [
          {
            "timestamp": "2025-04-26T06:00:00Z",
            "action": "edit_post"
          }
        ]
      },
      {
        "id": 10,
        "name": "Jack",
        "email": "jack@example.com",
        "active": true,
        "roles": [
          "user"
        ],
        "settings": {
          "theme": "light",
          "language": "ru"
        },
        "preferences": {
          "newsletter": true,
          "notifications": {
            "email": false,
            "sms": true
          }
        },
        "activity": [
          {
            "timestamp": "2025-04-25T15:00:00Z",
            "action": "create_post"
          }
        ]
      }
    ]
  }
}`
