package childprocess

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func FindProcessPath(process string) (string,error){
	bytes,_ := exec.Command("where.exe",process).Output()
	paths := strings.Split(string(bytes), "\n")
	pathss := strings.Split(paths[0], "\r")
	return pathss[0],nil
}

func findLineEnd(dat []byte) (out [][]byte) {
	prev := 0
	for pos, i := range dat {
		if i == []byte("\n")[0] {
			out = append(out, dat[prev:pos])
			prev = pos + 1
		}
	}

	out = append(out, dat[prev:])
	// for pos,i := range out {
	// 	count := 0;
	// 	for _,char := range i {
	// 		if (char == []byte(" ")[0]) {
	// 			count++;
	// 		}
	// 	}
	// 	if count == len(i)  && pos > 0{
	// 		out = append(out[:pos-1],out[pos:]...)
	// 	}
	// }
	return
}

func copyAndCapture(process string, w io.Writer, r io.Reader) {
	prefix := []byte(fmt.Sprintf("Child process (%s):", process))
	after := []byte("\n")
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			lines := findLineEnd(d)
			for _, line := range lines {
				out := append(prefix, line...)
				out = append(out, after...)

				_, err := w.Write(out)
				if err != nil {
					return
				}
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return
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