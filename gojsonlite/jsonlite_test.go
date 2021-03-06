// Copyright (c) 2019 Siemens AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// Author(s): Jonas Plum

package gojsonlite

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var (
	ExampleStore  = "example1.forensicstore"
	ProcessItemId = "process--920d7c41-0fef-4cf8-bce2-ead120f6b506"
	ProcessItem   = Item{
		"uid":          ProcessItemId,
		"artifact":     "IPTablesRules",
		"type":         "process",
		"name":         "iptables",
		"created":      "2016-01-20T14:11:25.550Z",
		"cwd":          "/root/",
		"arguments":    []interface{}{"-L", "-n", "-v"},
		"command_line": "/sbin/iptables -L -n -v",
		"stdout_path":  "IPTablesRules/stdout",
		"stderr_path":  "IPTablesRules/stderr",
		"return_code":  float64(0),
	}
)

func setup(t *testing.T) string {
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join("..", "test", "forensicstore", ExampleStore)
	dst := filepath.Join(dir, ExampleStore)
	for _, name := range []string{"item.db", filepath.Join("WindowsAMCacheHveFile", "Amcache.hve"), filepath.Join("IPTablesRules", "stderr"), filepath.Join("IPTablesRules", "stdout"), filepath.Join("WMILogicalDisks", "stdout"), filepath.Join("WMILogicalDisks", "wmi"), filepath.Join("WMILogicalDisks", "stderr")} {
		CopyFile(t, filepath.Join(src, name), filepath.Join(dst, name))
	}

	return dir + string(filepath.Separator)
}

func CopyFile(t *testing.T, src, dst string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(dst, input, 0644); err != nil {
		t.Fatal(err)
	}
}

func teardown(t *testing.T) {
	files, err := ioutil.ReadDir(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), t.Name()) {
			os.Remove(f.Name())
		}
	}
}

func TestNew(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type args struct {
		remoteUrl     string
		discriminator string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Create store", args{testDir + "new.jsonlite", "type"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.remoteUrl, tt.args.discriminator)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			os.Remove(tt.args.remoteUrl)
		})
	}
}

func TestJSONLite_Insert(t *testing.T) {
	foo := Item{"name": "foo", "type": "fo", "int": 0}
	bar := Item{"name": "bar", "type": "ba", "int": 2}
	baz := Item{"name": "baz", "type": "ba", "float": 0.1}
	bat := Item{"name": "bat", "type": "ba", "list": []string{}}
	bau := Item{"name": "bau", "type": "ba", "list": nil}

	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		item Item
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"Insert First", fields{testDir + ExampleStore}, args{foo}, "fo--", false},
		{"Insert Second", fields{testDir + ExampleStore}, args{bar}, "ba--", false},
		{"Insert Different Columns", fields{testDir + ExampleStore}, args{baz}, "ba--", false},
		{"Insert Empty List", fields{testDir + ExampleStore}, args{bat}, "ba--", false},
		{"Insert Item with nil", fields{testDir + ExampleStore}, args{bau}, "ba--", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			got, err := db.Insert(tt.args.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.Insert() error = %v, wantErr %v", err, tt.wantErr)
			} else if got[:4] != tt.want {
				t.Errorf("JSONLite.Insert() = %v, want %v", got[:4], tt.want)
			}
		})
	}
}

func TestJSONLite_Get(t *testing.T) {
	log.Print("get")
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantItem Item
		wantErr  bool
	}{
		{"Get item", fields{testDir + ExampleStore}, args{ProcessItemId}, ProcessItem, false},
		{"Get non existing", fields{testDir + ExampleStore}, args{"process--16b02a2b-d1a1-4e79-aad6-2f2c1c286818"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Print("get2")
			db, err := New(tt.fields.url, "type")
			log.Print("get3")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			// defer os.Remove(tt.fields.url)
			gotItem, err := db.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, gotItem, tt.wantItem)
		})
	}
}

func TestJSONLite_QueryJSONLite(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		query string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantItems []Item
		wantErr   bool
	}{
		{"Query", fields{testDir + ExampleStore}, args{"SELECT * FROM process WHERE name=\"iptables\""}, []Item{ProcessItem}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			gotItems, err := db.Query(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, gotItems, tt.wantItems)
		})
	}
}

/*
func TestJSONLite_Update(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	updatedItem := PROCESS_ITEM
	updatedItem["name"] = "foo"

	type fields struct {
		url string
	}
	type args struct {
		id          string
		partialItem Item
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"Update", fields{testDir + EXAMPLE_STORE}, args{PROCESS_ITEM_ID, Item{"name": "foo"}}, PROCESS_ITEM_ID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url,  "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			got, err := db.Update(tt.args.id, tt.args.partialItem)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSONLite.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/

func TestJSONLite_Discriminator(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Discriminator", fields{testDir + ExampleStore}, "type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			if got := db.Discriminator(); got != tt.want {
				t.Errorf("JSONLite.Discriminator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONLite_SetDiscriminator(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		discriminator string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Set Discriminator", fields{testDir + ExampleStore}, args{"type"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			db.SetDiscriminator(tt.args.discriminator)
			assert.Equal(t, tt.args.discriminator, db.Discriminator())
		})
	}
}

func TestJSONLite_Strict(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Strict", fields{testDir + ExampleStore}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			if got := db.Strict(); got != tt.want {
				t.Errorf("JSONLite.Strict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONLite_SetStrict(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		strict bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Discriminator", fields{testDir + ExampleStore}, args{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			db.SetStrict(tt.args.strict)
			assert.Equal(t, tt.args.strict, db.Strict())
		})
	}
}

func TestJSONLite_Select(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		itemType   string
		conditions []map[string]string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantItems int
		wantErr   bool
	}{
		{"Select", fields{testDir + ExampleStore}, args{"file", nil}, 2, false},
		{"Select with filter", fields{testDir + ExampleStore}, args{"file", []map[string]string{{"name": "foo.doc"}}}, 1, false},
		{"Select not existing", fields{testDir + ExampleStore}, args{"xxx", nil}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			gotItems, err := db.Select(tt.args.itemType, tt.args.conditions)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// assert.EqualValues(t, gotItems, tt.wantItems) // TODO check array content
			assert.EqualValues(t, tt.wantItems, len(gotItems))
		})
	}
}

func TestJSONLite_All(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	tests := []struct {
		name      string
		fields    fields
		wantItems int
		wantErr   bool
	}{
		{"All", fields{testDir + ExampleStore}, 7, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			gotItems, err := db.All()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// assert.EqualValues(t, gotItems, tt.wantItems)
			assert.Equal(t, tt.wantItems, len(gotItems))
		})
	}
}

/*
type MockColumnType struct {
	sql.ColumnType
	t reflect.Type
}

func (ci *MockColumnType) ScanType() reflect.Type {
	return ci.t
}

type MockRows struct {
	*sql.Rows
	i     int
	items []Item
}

func NewMockRows() *MockRows {
	rs := MockRows{}
	rs.i = 2
	rs.items = []Item{
		{"id": 1, "foo": map[string]interface{}{"bar": "post"}, "body": "hello"},
		{"id": 2, "foo": map[string]interface{}{"bar": "man"}, "body": "world"},
	}
	return &rs
}

func (rs *MockRows) Next() bool {
	rs.i--
	return rs.i > 0
}
func (rs *MockRows) Scan(dest ...interface{}) error {
	dest[0] = rs.items[0]["id"]
	dest[0] = rs.items[0]["foo"]
	dest[0] = rs.items[0]["body"]
	return nil
}
func (rs *MockRows) ColumnTypeScanType(index int) reflect.Type {
	columns := []reflect.Type{
		reflect.TypeOf(1),
		reflect.TypeOf(map[string]interface{}{}),
		reflect.TypeOf(""),
	}

	return columns[index]
}
*/

func TestJSONLite_rowsToItems(t *testing.T) {

	/* TODO create MockRows


	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantItems []Item
		wantErr   bool
	}{
		{"Row to Items", fields{testDir + EXAMPLE_STORE}, args{rows}, items, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url,  "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			gotItems, err := db.rowsToItems(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.rowsToItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, gotItems, tt.wantItems)
		})
	}
	*/
}

func TestJSONLite_getTables(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	expectedTables := map[string]map[string]string{
		"directory":            {"accessed": "TEXT", "artifact": "TEXT", "created": "TEXT", "modified": "TEXT", "path": "TEXT", "type": "TEXT", "uid": "TEXT"},
		"file":                 {"accessed": "TEXT", "artifact": "TEXT", "created": "TEXT", "export_path": "TEXT", "hashes.MD5": "TEXT", "hashes.SHA-1": "TEXT", "modified": "TEXT", "name": "TEXT", "origin.path": "TEXT", "origin.volume": "TEXT", "size": "INTEGER", "type": "TEXT", "uid": "TEXT"},
		"process":              {"arguments.0": "TEXT", "arguments.1": "TEXT", "arguments.2": "TEXT", "artifact": "TEXT", "command_line": "TEXT", "created": "TEXT", "cwd": "TEXT", "name": "TEXT", "return_code": "INTEGER", "stderr_path": "TEXT", "stdout_path": "TEXT", "type": "TEXT", "uid": "TEXT", "wmi_path": "TEXT"},
		"windows-registry-key": {"artifact": "TEXT", "key": "TEXT", "modified": "TEXT", "type": "TEXT", "uid": "TEXT", "values.0.data": "TEXT", "values.0.data_type": "TEXT", "values.0.name": "TEXT"},
	}

	type fields struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]map[string]string
		wantErr bool
	}{
		{"Get Tables", fields{testDir + ExampleStore}, expectedTables, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			got, err := db.getTables()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.getTables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSONLite.getTables() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestJSONLite_ensureTable(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		flatItem Item
		item     Item
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Ensure table", fields{testDir + ExampleStore}, args{Item{"foo": 1, "type": "bar"}, Item{"foo": 1, "type": "bar"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			if err := db.ensureTable(tt.args.flatItem, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.ensureTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONLite_createTable(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		flatItem Item
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Create table", fields{testDir + ExampleStore}, args{Item{"foo": 1, "type": "bar"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			if err := db.createTable(tt.args.flatItem); (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.createTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getSQLDataType(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Get SQL Data Type INTEGER", args{1}, "INTEGER"},
		{"Get SQL Data Type TEXT", args{"foo"}, "TEXT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSQLDataType(tt.args.value); got != tt.want {
				t.Errorf("getSQLDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONLite_createOptionsTable(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Create Option Table", fields{testDir + ExampleStore}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)
			if err := db.createOptionsTable(); (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.createOptionsTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONLite_option(t *testing.T) {
	testDir := setup(t)
	defer teardown(t)

	type fields struct {
		url string
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{"Option int", fields{testDir + ExampleStore}, args{"foo", 1}, 1, false},
		{"Option string", fields{testDir + ExampleStore}, args{"foo", "bar"}, "bar", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)

			if err := db.setOption(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.setOption() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := db.option(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.option() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.args.value) {
				t.Errorf("JSONLite.option() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONLite_Validate(t *testing.T) {
	testDir := setup(t)

	type fields struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		wantE   []string
		wantErr bool
	}{
		{"Validate valid", fields{testDir + ExampleStore}, []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.fields.url, "type")
			if err != nil || db == nil {
				t.Fatalf("Database could not be created %v\n", err)
			}

			defer os.Remove(tt.fields.url)

			gotE, err := db.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLite.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotE, tt.wantE) {
				t.Errorf("JSONLite.Validate() = %v, want %v", gotE, tt.wantE)
			}
		})
	}
}
