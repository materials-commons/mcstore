package mc

type mcprojects struct {

}

var Projects mcprojects = mcprojects{

}

func (p *mcprojects) All() ([]ProjectDB, error) {
	return nil, nil
}