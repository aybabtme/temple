package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

var tests = []struct {
	Template string
	Flags    []string
	Want     string
}{
	{
		Template: `{"hello": "{{.hello}}"}`,
		Flags:    []string{"-var", `hello=world`},
		Want:     `{"hello": "world"}`,
	},
}

func TestTemple(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "temple-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := exec.Command("go", "build", "-o", dir+"/tmpl.bin", "main.go").Run(); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		in := bytes.NewBufferString(tt.Template)
		vars := tt.Flags
		want := bytes.NewBufferString(tt.Want)

		stderr := bytes.NewBuffer(nil)
		got := bytes.NewBuffer(nil)

		cmd := exec.Command(dir+"/tmpl.bin", vars...)
		cmd.Stdin = in
		cmd.Stdout = got
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			t.Error(stderr.String())
			t.Fatal(err)
		}

		if want.String() != got.String() {
			t.Errorf("want=%q", want.String())
			t.Errorf(" got=%q", got.String())
		}
	}
}
