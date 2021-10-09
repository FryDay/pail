package pail

import "github.com/FryDay/pail/sqlite"

type Var struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type Value struct {
	ID    int64  `db:"id"`
	VarID int64  `db:"var_id"`
	Value string `db:"value"`
}

type VarValue struct {
	*Var
	*Value
}

func NewVar(name string) *Var {
	return &Var{Name: name}
}

func NewValue(varID int64, value string) *Value {
	return &Value{VarID: varID, Value: value}
}

func getVar(db *sqlite.DB, name string) (*Var, error) {
	v := &Var{}
	if err := db.Get(`select id, name from var where name=:name`, map[string]interface{}{"name": name}, v); err != nil {
		if err == sqlite.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return v, nil
}

func getValue(db *sqlite.DB, varID int64, value string) (*Value, error) {
	val := &Value{}
	if err := db.Get(`select id, var_id, value from value where var_id=:var_id and value=:value`, map[string]interface{}{"var_id": varID, "value": value}, val); err != nil {
		if err == sqlite.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func getVarValue(db *sqlite.DB, name string) (*VarValue, error) {
	varValue := &VarValue{}
	if err := db.Get(`select ('$' || v.name) name, val.value from value val join var v on v.id = val.var_id where v.name=:name order by random() limit 1`, map[string]interface{}{"name": name}, varValue); err != nil {
		if err == sqlite.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return varValue, nil
}

func (v *Var) insert(db *sqlite.DB) (err error) {
	v.ID, err = db.NamedExec(`insert into var (name) values (:name)`, v)
	return err
}

func (val *Value) insert(db *sqlite.DB) (err error) {
	val.ID, err = db.NamedExec(`insert into value (var_id, value) values (:var_id, :value)`, val)
	return err
}
