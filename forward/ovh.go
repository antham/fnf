package forward

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ovh/go-ovh/ovh"
)

type ovhClient interface {
	Get(string, interface{}) error
	Post(string, interface{}, interface{}) error
	Delete(string, interface{}) error
}

type createRedirectionRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	LocalCopy bool   `json:"localCopy"`
}

type ForwardInfo struct {
	From string `json:"from"`
	ID   string `json:"id,omitempty"`
	To   string `json:"to"`
}

type OVHProvider struct {
	client       ovhClient
	domain       string
	defaultEmail string
}

func NewOVHProvider(endpoint string, appKey string, appSecret string, consumerKey string, domain string, defaultEmail string) (*OVHProvider, error) {
	client, err := ovh.NewClient(
		endpoint,
		appKey,
		appSecret,
		consumerKey,
	)
	if err != nil {
		return nil, err
	}
	return &OVHProvider{client: client, domain: domain, defaultEmail: defaultEmail}, nil
}

func (o *OVHProvider) Create(localPartFrom string, emailTo string) error {
	return o.client.Post(
		fmt.Sprintf("/email/domain/%s/redirection", o.domain),
		createRedirectionRequest{
			From: fmt.Sprintf("%s@%s", localPartFrom, o.domain),
			To:   emailTo,
		}, nil)
}

func (o *OVHProvider) CreateOnDefaultEmail() error {
	return o.Create(uuid.New().String(), o.defaultEmail)
}

func (o *OVHProvider) List() ([]ForwardInfo, error) {
	var ids []string
	err := o.client.Get(fmt.Sprintf("/email/domain/%s/redirection", o.domain), &ids)
	if err != nil {
		return []ForwardInfo{}, err
	}
	infos := []ForwardInfo{}
	for _, id := range ids {
		info := ForwardInfo{}
		err := o.client.Get(fmt.Sprintf("/email/domain/%s/redirection/%s", o.domain, id), &info)
		if err != nil {
			return []ForwardInfo{}, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (o *OVHProvider) Delete(id string) error {
	return o.client.Delete(fmt.Sprintf("/email/domain/%s/redirection/%s", o.domain, id), nil)
}
