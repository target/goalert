package swosync

type (
	RowSet map[RowID]struct{}
	RowID  struct{ Table, Row string }
)

func (r RowSet) Set(id RowID)    { r[id] = struct{}{} }
func (r RowSet) Delete(id RowID) { delete(r, id) }

func (r RowSet) Has(id RowID) bool {
	_, ok := r[id]
	return ok
}
