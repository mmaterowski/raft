package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteToFile(id string, filename string, payload string) {
	data := []byte(payload)
	strings.Title(id)
	path := "persistence/" + strings.Title(id) + "/" + filename
	err := ioutil.WriteFile(path, data, 0644)
	check(err)
}
