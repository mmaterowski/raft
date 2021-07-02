package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func WriteToFile(id string, payload string) {
	data := []byte(payload)
	path := "persistence/" + string(id)
	err := ioutil.WriteFile(path, data, 0644)
	check(err)
}
