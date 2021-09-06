package db

const (
	ConstLayoutDateTime  = `2006-01-02 15:04`
	ConstLayoutDateTimeZ = `T2006-01-02Z 15:04:00`
	ConstLayoutDate      = `2006-01-02`
	ConstLayoutTime      = `15:04`
)

var ConstRoles = struct {
	Admin    int
	Cashier  int
	Reseller int
	Client   int
	API      int
}{
	Admin:    1,
	Cashier:  2,
	Reseller: 3,
	Client:   4,
	API:      5,
}
