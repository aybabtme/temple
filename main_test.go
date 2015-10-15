package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTempleFile(t *testing.T) {

	withBuild(t, func(biname string) {
		dirs, err := ioutil.ReadDir("testdata/file")
		if err != nil {
			panic(err)
		}

		for _, tt := range dirs {
			if !tt.IsDir() {
				continue
			}
			t.Logf("test %q", tt.Name())
			flagsPath := filepath.Join("testdata/file", tt.Name(), "flags")
			goldPath := filepath.Join("testdata/file", tt.Name(), "gold.json")
			srcPath := filepath.Join("testdata/file", tt.Name(), "src.tpl.json")

			flags, err := ioutil.ReadFile(flagsPath)
			if err != nil {
				panic(err)
			}

			gold, err := ioutil.ReadFile(goldPath)
			if err != nil {
				panic(err)
			}

			tpl, err := ioutil.ReadFile(srcPath)
			if err != nil {
				panic(err)
			}

			in := bytes.NewBuffer(tpl)
			vars := append(
				[]string{"file"},
				strings.Split(strings.TrimSpace(string(flags)), " ")...,
			)
			want := bytes.NewBuffer(gold)

			stderr := bytes.NewBuffer(nil)
			got := bytes.NewBuffer(nil)

			cmd := exec.Command(biname, vars...)
			cmd.Stdin = in
			cmd.Stdout = got
			cmd.Stderr = stderr
			if err := cmd.Run(); err != nil {
				t.Error(stderr.String())
				panic(err)
			}

			if want.String() != got.String() {
				t.Errorf("want=%q", want.String())
				t.Errorf(" got=%q", got.String())
			}
		}
	})
}

func withBuild(t testing.TB, test func(binname string)) {
	dir, err := ioutil.TempDir(os.TempDir(), "temple-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	if err := exec.Command("go", "build", "-o", dir+"/tmpl.bin", "main.go").Run(); err != nil {
		panic(err)
	}
	test(dir + "/tmpl.bin")
}

func TestTempleTree(t *testing.T) {

	withBuild(t, func(biname string) {
		dirs, err := ioutil.ReadDir("testdata/tree")
		if err != nil {
			panic(err)
		}

		for _, tt := range dirs {
			if !tt.IsDir() {
				continue
			}
			t.Logf("test %q", tt.Name())
			flagsPath := filepath.Join("testdata/tree", tt.Name(), "flags")
			goldPath := filepath.Join("testdata/tree", tt.Name(), "gold")
			srcPath := filepath.Join("testdata/tree", tt.Name(), "src")
			dstPath := filepath.Join("testdata/tree", tt.Name(), "dst")

			flags, err := ioutil.ReadFile(flagsPath)
			if err != nil {
				panic(err)
			}

			withTmpDir(t, func(tmpDir string) {
				copyDir(t, dstPath, tmpDir)

				vars := append(
					[]string{"tree", "-src", srcPath, "-dst", tmpDir},
					strings.Split(strings.TrimSpace(string(flags)), " ")...,
				)

				stderr := bytes.NewBuffer(nil)
				cmd := exec.Command(biname, vars...)
				cmd.Stderr = stderr
				if err := cmd.Run(); err != nil {
					t.Error(stderr.String())
					panic(err)
				}

				verifyEqual(t, goldPath, tmpDir)
			})
		}
	})
}

func copyDir(t testing.TB, src, dst string) {
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if fi.IsDir() {
			return nil
		}
		t.Logf("copying file %q", path)

		dirStat, err := os.Stat(filepath.Dir(path))
		if err != nil {
			panic(err)
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			panic(err)
		}

		tgt := filepath.Join(dst, rel)

		mkdir := filepath.Dir(tgt)
		if err := os.MkdirAll(mkdir, dirStat.Mode().Perm()); err != nil && !os.IsExist(err) {
			panic(err)
		}
		t.Logf("created dir %q", mkdir)

		srcFile, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(tgt)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			panic(err)
		}
		if err := dstFile.Close(); err != nil {
			panic(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func verifyEqual(t testing.TB, gold, got string) {
	err := filepath.Walk(gold, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == gold {
			return nil
		}
		rel, err := filepath.Rel(gold, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(got, rel)

		// t.Logf("\ngold=%q\npath=%q\nrel=%q\ndst=%q", gold, path, rel, dstPath)

		wantStat, err := os.Stat(path)
		if err != nil {
			return err
		}
		gotStat, err := os.Stat(dstPath)
		if os.IsNotExist(err) {
			t.Errorf("file %q should exist, but it doesn't", dstPath)
			return nil
		} else if err != nil {
			return err
		}

		if want, got := wantStat.Mode().Perm(), gotStat.Mode().Perm(); want != got {
			t.Errorf("%q: want perm %v, got %v", dstPath, want, got)
		}

		if wantStat.IsDir() && gotStat.IsDir() {
			return nil
		}
		if wantStat.IsDir() != gotStat.IsDir() {
			t.Errorf("isDir=%v: %q", wantStat.IsDir(), wantStat.Name())
			t.Errorf("isDir=%v: %q", gotStat.IsDir(), gotStat.Name())
			return nil
		}

		wantBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		gotBytes, err := ioutil.ReadFile(dstPath)
		if os.IsNotExist(err) {
			t.Errorf("file %q should exist", dstPath)
			return nil
		} else if err != nil {
			return err
		}

		if want, got := wantBytes, gotBytes; !bytes.Equal(want, got) {
			t.Errorf("%q: want=%q", path, string(want))
			t.Errorf("%q:  got=%q", path, string(got))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(got, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(got, path)
		if err != nil {
			return err
		}
		if _, err := os.Stat(filepath.Join(gold, rel)); os.IsNotExist(err) {
			t.Errorf("should not exist on target: %v", path)
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func withTmpDir(t testing.TB, test func(tmpdir string)) {
	dir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("temple-test-%d", rand.Int()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	test(dir)
}
