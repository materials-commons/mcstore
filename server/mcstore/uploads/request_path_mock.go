package uploads

import "github.com/materials-commons/mcstore/pkg/app/flow"

type mockRequestPath struct {
	mcdirRP *mcdirRequestPath
	err     error
}

func (p *mockRequestPath) path(req *flow.Request) string {
	return p.mcdirRP.path(req)
}

func (p *mockRequestPath) dir(req *flow.Request) string {
	return p.mcdirRP.dir(req)
}

func (p *mockRequestPath) dirFromID(id string) string {
	return p.mcdirRP.dirFromID(id)
}

func (p *mockRequestPath) mkdir(req *flow.Request) error {
	return p.err
}

func (p *mockRequestPath) mkdirFromID(id string) error {
	return p.err
}
