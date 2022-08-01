package swosync

type (
	rowSet   map[changeID]struct{}
	changeID struct{ Table, Row string }
)

func (r rowSet) Set(id changeID)    { r[id] = struct{}{} }
func (r rowSet) Delete(id changeID) { delete(r, id) }

func (r rowSet) Has(id changeID) bool {
	_, ok := r[id]
	return ok
}
