package user

// FakeInserter is a test fake for dbaccess.Inserter[Record].
type FakeInserter struct{ Err error }

// Insert implements the dbaccess.Inserter[Record] interface on FakeInserter.
func (f *FakeInserter) Insert(_ Record) error { return f.Err }

// FakeSelector is a test fake for dbaccess.Selector[Record].
type FakeSelector struct {
	User Record
	Err  error
}

// Select implements the dbaccess.Selector[Record] interface on FakeSelector.
func (f *FakeSelector) Select(_ string) (Record, error) {
	return f.User, f.Err
}
