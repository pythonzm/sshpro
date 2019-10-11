package copy

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"os"
	"path/filepath"
	"sshpro/utils/root"
	"strings"
	"sync"
)

const Separator = string(os.PathSeparator)

type Cp struct {
	root.Result
	LocalFilePath  string
	RemoteFilePath string
}

var dirSlice, fileSlice []string
var basep string

//单个copy
func Scp(c Cp, ch chan bool, wg *sync.WaitGroup) (output string, err error) {
	r := root.Result{Connection: c.Connection}
	defer wg.Done()
	defer func() { <-ch }()
	ch <- true

	o, b := c.checkIfDir()
	if b {
		return o, nil
	}
	sftpClient, e := c.SFTPConnect()

	if e != nil {
		err = c.genError(e)
		return
	}
	defer func() {
		if e := sftpClient.Close(); e != nil {
			return
		}
	}()

	var remoteFileName string
	info, e := sftpClient.Stat(c.RemoteFilePath)

	if e == nil {
		if info.IsDir() {
			remoteFileName = filepath.Join(c.RemoteFilePath, filepath.Base(c.LocalFilePath))
		} else {
			remoteFileName = c.RemoteFilePath
		}
	} else {
		remoteFileName = c.RemoteFilePath
	}

	if err = cpFile(sftpClient, c.LocalFilePath, remoteFileName); err != nil {
		err = c.genError(err)
		return
	}
	r.Msg = fmt.Sprintf("%s 传输完成\n", c.LocalFilePath)
	r.Status = r.ResultStatus(0)
	output = r.ColorResult()
	return
}

func MultiScp(c Cp, ch chan bool, wg *sync.WaitGroup) (output string, err error) {
	o, b := c.checkIfDir()
	if !b {
		return o, nil
	}

	if strings.HasSuffix(c.LocalFilePath, Separator) {
		strings.TrimRight(c.LocalFilePath, Separator)
	}
	basep = c.LocalFilePath

	var remoteBasePath string

	if strings.HasSuffix(c.RemoteFilePath, "/") {
		remoteBasePath = filepath.Join(c.RemoteFilePath, filepath.Base(basep))
	} else {
		remoteBasePath = c.RemoteFilePath
	}

	r := root.Result{Connection: c.Connection}

	sftpClient, e := c.SFTPConnect()
	if e != nil {
		err = c.genError(e)
		return
	}
	defer func() {
		if e := sftpClient.Close(); e != nil {
			return
		}
	}()

	dirSlice = dirSlice[:0]
	fileSlice = fileSlice[:0]
	if e := filepath.Walk(c.LocalFilePath, walkFunc); e != nil {
		return
	}

	_, err = sftpClient.Stat(remoteBasePath)
	if err == nil {
		err = errors.New(fmt.Sprintf("目标目录%s已存在", remoteBasePath))
		r = r.GenResult(err)
		err = errors.New(r.ColorResult())
		return
	}
	if err = sftpClient.Mkdir(remoteBasePath); err != nil {
		return
	}

	for _, value := range dirSlice {

		err = cpDir(sftpClient, filepath.Join(remoteBasePath, value))
		if err != nil {
			return
		}
	}

	for _, value := range fileSlice {
		wg.Add(1)
		go func(v string) {
			ch <- true
			localFile := filepath.Join(c.LocalFilePath, v)
			remoteFile := filepath.Join(remoteBasePath, v)

			if e := cpFile(sftpClient, localFile, remoteFile); e != nil {
				r = r.GenResult(e)
				err = errors.New(r.ColorResult())
				return
			}
			<-ch
			wg.Done()
		}(value)
	}
	wg.Wait()

	r.Msg = fmt.Sprintf("%s 传输完成\n", c.LocalFilePath)
	r.Status = r.ResultStatus(0)
	output = r.ColorResult()
	return
}

func walkFunc(p string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if p != basep {
		rel, e := filepath.Rel(basep, p)
		if e != nil {
			return e
		}
		if info.IsDir() {
			dirSlice = append(dirSlice, rel)
		} else {
			fileSlice = append(fileSlice, rel)
		}
	}

	return nil
}

func cpFile(client *sftp.Client, localFile, remoteFile string) error {

	srcFile, e := os.Open(localFile)
	if e != nil {
		return e
	}

	defer func() {
		if e := srcFile.Close(); e != nil {
			return
		}
	}()

	dstFile, e := client.Create(remoteFile)
	if e != nil {
		return e
	}
	defer func() {
		if e := dstFile.Close(); e != nil {
			return
		}
	}()

	buf := make([]byte, 1024*2)

	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		_, e := dstFile.Write(buf[0:n])
		if e != nil {
			return e
		}
	}
	return nil
}

func cpDir(client *sftp.Client, remotePath string) error {
	if e := client.Mkdir(remotePath); e != nil {
		return e
	}
	return nil
}

func (c Cp) checkIfDir() (string, bool) {
	r := root.Result{Connection: c.Connection}
	localFileInfo, e := os.Stat(c.LocalFilePath)
	if e != nil {
		r.Status = r.ResultStatus(1)
		r.Msg = e.Error()
		output := r.ColorResult()
		return output, false
	}
	if localFileInfo.IsDir() {
		r.Status = r.ResultStatus(1)
		r.Msg = fmt.Sprintf("%s是文件夹，请使用-r参数\n", localFileInfo.Name())
		output := r.ColorResult()
		return output, true
	} else {
		r.Status = r.ResultStatus(1)
		r.Msg = fmt.Sprintf("%s是文件，请去掉-r参数\n", localFileInfo.Name())
		output := r.ColorResult()
		return output, false
	}
}

func (c Cp) genError(e error) error {
	r := root.Result{Connection: c.Connection}
	r = r.GenResult(e)
	err := errors.New(r.ColorResult())
	return err
}

