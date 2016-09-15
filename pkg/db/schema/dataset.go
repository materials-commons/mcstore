package schema

type Dataset struct {
	ID        string `gorethink:"id,omitempty" json:"id"`
	Published bool   `gorethink:"published" json:"published"`
}
