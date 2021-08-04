package helpers

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
)

func Check(e error) {
	if e != nil {
		log.Print(e)
		panic(e)
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func PrintAsciiHelloString() {
	log.Print(`
    Raft implementation by Michal Materowski, 2k21
	.----------------.  .----------------.  .----------------.  .----------------. 
	| .--------------. || .--------------. || .--------------. || .--------------. |
	| |  _______     | || |      __      | || |  _________   | || |  _________   | |
	| | |_   __ \    | || |     /  \     | || | |_   ___  |  | || | |  _   _  |  | |
	| |   | |__) |   | || |    / /\ \    | || |   | |_  \_|  | || | |_/ | | \_|  | |
	| |   |  __ /    | || |   / ____ \   | || |   |  _|      | || |     | |      | |
	| |  _| |  \ \_  | || | _/ /    \ \_ | || |  _| |_       | || |    _| |_     | |
	| | |____| |___| | || ||____|  |____|| || | |_____|      | || |   |_____|    | |
	| |              | || |              | || |              | || |              | |
	| '--------------' || '--------------' || '--------------' || '--------------' |
	 '----------------'  '----------------'  '----------------'  '----------------' 
	 
								 
  `)
}
