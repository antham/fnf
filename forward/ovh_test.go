package forward

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOVHProvider(t *testing.T) {
	provider, err := NewOVHProvider("ovh-eu", "ABCD", "ABCD", "ABCD", "test.xyz", "test@test.com")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreate(t *testing.T) {
	type testCase struct {
		name             string
		localPartFromArg string
		emailToArg       string
		setup            func(*testing.T) *mockOvhClient
		test             func(*testing.T, error)
	}
	for _, tt := range []testCase{
		{
			"An error occurred when creating a redirection",
			"order",
			"whatever@test.com",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Post",
					"/email/domain/test.xyz/redirection",
					createRedirectionRequest{
						From: "order@test.xyz",
						To:   "whatever@test.com",
					},
					nil,
				).Return(errors.New("an error occurred"))
				return client
			},
			func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			"Create a redirection",
			"order",
			"whatever@test.com",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Post",
					"/email/domain/test.xyz/redirection",
					createRedirectionRequest{
						From: "order@test.xyz",
						To:   "whatever@test.com",
					},
					nil,
				).Return(nil)
				return client
			},
			func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setup(t)
			provider := &OVHProvider{
				client: client,
				domain: "test.xyz",
			}
			tt.test(t, provider.Create(tt.localPartFromArg, tt.emailToArg))
		})
	}
}

func TestCreateOnDefaultEmail(t *testing.T) {
	type testCase struct {
		name  string
		setup func(*testing.T) *mockOvhClient
		test  func(*testing.T, error)
	}
	for _, tt := range []testCase{
		{
			"An error occurred when creating a redirection on default email",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Post",
					"/email/domain/test.xyz/redirection",
					mock.MatchedBy(func(r createRedirectionRequest) bool {
						assert.Equal(t, "whatever@test.com", r.To)
						s := strings.Split(r.From, "@")
						_, err := uuid.Parse(s[0])
						assert.NoError(t, err)
						assert.Equal(t, "test.xyz", s[1])
						return true
					}),
					nil,
				).Return(errors.New("an error occurred"))
				return client
			},
			func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			"Create a redirection on default email",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Post",
					"/email/domain/test.xyz/redirection",
					mock.MatchedBy(func(r createRedirectionRequest) bool {
						assert.Equal(t, "whatever@test.com", r.To)
						s := strings.Split(r.From, "@")
						_, err := uuid.Parse(s[0])
						assert.NoError(t, err)
						assert.Equal(t, "test.xyz", s[1])
						return true
					}),
					nil,
				).Return(nil)
				return client
			},
			func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setup(t)
			provider := &OVHProvider{
				client:       client,
				domain:       "test.xyz",
				defaultEmail: "whatever@test.com",
			}
			tt.test(t, provider.CreateOnDefaultEmail())
		})
	}
}

func TestList(t *testing.T) {
	type testCase struct {
		name  string
		setup func(*testing.T) *mockOvhClient
		test  func(*testing.T, []ForwardInfo, error)
	}
	for _, tt := range []testCase{
		{
			"An error occurred when fetching the redirection info",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Get",
					"/email/domain/test.xyz/redirection",
					mock.Anything,
				).Return(errors.New("an error occurred"))
				return client
			},
			func(t *testing.T, f []ForwardInfo, err error) {
				assert.Error(t, err)
			},
		},
		{
			"An error occurred when fetching the redirection info",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Get",
					"/email/domain/test.xyz/redirection",
					mock.MatchedBy(func(ids *[]string) bool {
						*ids = append(*ids, "d98c97a5-ca2c-45b6-978b-dbec5e4bae70")
						return true
					}),
				).Return(nil)
				client.Mock.On(
					"Get",
					"/email/domain/test.xyz/redirection/d98c97a5-ca2c-45b6-978b-dbec5e4bae70",
					mock.Anything,
				).Return(errors.New("an error occurred"))
				return client
			},
			func(t *testing.T, f []ForwardInfo, err error) {
				assert.Error(t, err)
			},
		},
		{
			"Returns a detail of all ongoing redirection",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Get",
					"/email/domain/test.xyz/redirection",
					mock.MatchedBy(func(ids *[]string) bool {
						*ids = append(*ids, "fc2b86af-1f31-4c9b-97ec-f968f06498d2", "7dc3f818-ff13-433f-bd00-958e68572a46")
						return true
					}),
				).Return(nil)
				client.Mock.On(
					"Get",
					"/email/domain/test.xyz/redirection/fc2b86af-1f31-4c9b-97ec-f968f06498d2",
					mock.MatchedBy(func(info *ForwardInfo) bool {
						*info = ForwardInfo{
							From: "test@test.xyz",
							To:   "test@test.com",
							ID:   "fc2b86af-1f31-4c9b-97ec-f968f06498d2",
						}
						return true
					}),
				).Return(nil)
				client.Mock.On(
					"Get",
					"/email/domain/test.xyz/redirection/7dc3f818-ff13-433f-bd00-958e68572a46",
					mock.MatchedBy(func(info *ForwardInfo) bool {
						*info = ForwardInfo{
							From: "test1@test.xyz",
							To:   "test1@test.com",
							ID:   "7dc3f818-ff13-433f-bd00-958e68572a46",
						}
						return true
					}),
				).Return(nil)
				return client
			},
			func(t *testing.T, f []ForwardInfo, err error) {
				assert.NoError(t, err)
				assert.Equal(t, []ForwardInfo{
					{
						From: "test@test.xyz",
						ID:   "fc2b86af-1f31-4c9b-97ec-f968f06498d2",
						To:   "test@test.com",
					},
					{
						From: "test1@test.xyz",
						ID:   "7dc3f818-ff13-433f-bd00-958e68572a46",
						To:   "test1@test.com",
					},
				}, f)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setup(t)
			provider := &OVHProvider{
				client:       client,
				domain:       "test.xyz",
				defaultEmail: "whatever@test.com",
			}
			f, err := provider.List()
			tt.test(t, f, err)
		})
	}
}

func TestDelete(t *testing.T) {
	type testCase struct {
		name  string
		setup func(*testing.T) *mockOvhClient
		test  func(*testing.T, error)
	}
	for _, tt := range []testCase{
		{
			"An error occurred when deleting a redirection",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Delete",
					"/email/domain/test.xyz/redirection/bbefc810-2cfb-4cb7-b6ae-081467323fe3",
					nil,
				).Return(errors.New("an error occurred"))
				return client
			},
			func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			"Delete a redirection",
			func(t *testing.T) *mockOvhClient {
				client := newMockOvhClient(t)
				client.Mock.On(
					"Delete",
					"/email/domain/test.xyz/redirection/bbefc810-2cfb-4cb7-b6ae-081467323fe3",
					nil,
				).Return(nil)
				return client
			},
			func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setup(t)
			provider := &OVHProvider{
				client:       client,
				domain:       "test.xyz",
				defaultEmail: "whatever@test.com",
			}
			tt.test(t, provider.Delete("bbefc810-2cfb-4cb7-b6ae-081467323fe3"))
		})
	}
}
