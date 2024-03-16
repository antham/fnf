package forward

import (
	"fmt"

	"github.com/ovh/go-ovh/ovh"
)

type ForwardInfo struct {
	From string `json:"from"`
	ID   string `json:"id,omitempty"`
	To   string `json:"to"`
}

type Forward interface {
	Create(string) error
	List() ([]ForwardInfo, error)
	Delete(string) error
}

type OVHProvider struct {
	client *ovh.Client
	domain string
}

func NewOVHProvider(endpoint string, appKey string, appSecret string, consumerKey string, domain string) (*OVHProvider, error) {
	client, err := ovh.NewClient(
		endpoint,
		appKey,
		appSecret,
		consumerKey,
	)
	if err != nil {
		return nil, err
	}
	return &OVHProvider{client: client, domain: domain}, nil
}

func (o *OVHProvider) Create(name string) error {
	return o.client.Post(
		fmt.Sprintf("/email/domain/%s/redirection", o.domain),
		struct {
			From      string `json:"from"`
			To        string `json:"to"`
			LocalCopy bool   `json:"localCopy"`
		}{
			From:      fmt.Sprintf("%s@%s", name, o.domain),
			To:        name,
			LocalCopy: false,
		}, nil)
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
	return infos, err
}

func (o *OVHProvider) Delete(id string) error {
	return o.client.Delete(fmt.Sprintf("/email/domain/%s/redirection/%s", o.domain, id), nil)
}
