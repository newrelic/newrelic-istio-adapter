package internal

import "testing"

func TestCompress(t *testing.T) {
	input := "this is the input string that needs to be compressed"
	buf, err := Compress([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	back, err := Uncompress(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if string(back) != input {
		t.Error(string(back))
	}
}
