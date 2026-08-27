package model

import "testing"

func TestStringSliceDatabaseValueAndScan(t *testing.T) {
	value, err := StringSlice{"go", "remote"}.Value()
	if err != nil || string(value.([]byte)) != `["go","remote"]` {
		t.Fatalf("value = %#v, %v", value, err)
	}
	var scanned StringSlice
	if err := scanned.Scan(value.([]byte)); err != nil || len(scanned) != 2 || scanned[1] != "remote" {
		t.Fatalf("scan = %#v, %v", scanned, err)
	}
	if err := scanned.Scan("wrong type"); err == nil {
		t.Fatal("non-byte input was accepted")
	}
}
