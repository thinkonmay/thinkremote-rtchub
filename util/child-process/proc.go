package childprocess

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func FindProcessPath(process string) (string,error){
	first := strings.Split(os.Args[0],"\\");
	prefixs := first[:(len(first)-1)]
	prefixs = append(prefixs, process)
	merge := strings.Join(prefixs, "\\")
	return merge,nil
}


func copyAndCapture(process string, w io.Writer, r io.Reader) {
	procname := strings.Split(process,"\\")
	prefix := []byte(fmt.Sprintf("Child process (%s): ", procname[len(procname)-1]))
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return
		}


		if n < 1 {
			continue
		}


		lines := strings.Split(string(buf[:n]),"\n")
		for _,line := range lines {
			if len(line) < 2 {
				continue
			}
			_,err := w.Write([]byte(fmt.Sprintf("%s%s\n",prefix,line)))
			if err != nil {
				return
			}
		}
	}
}
func HandleProcess(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	processname := cmd.Args[0]
	fmt.Printf("handling process %s\n",processname)

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	cmd.Start()
	go func() {
		copyAndCapture(processname, os.Stdout, stdoutIn)
	}()
	go func() {
		copyAndCapture(processname, os.Stdout, stderrIn)
	}()
	cmd.Wait()

}