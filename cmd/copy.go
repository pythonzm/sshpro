/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	copy2 "sshpro/utils/copy"
	"sshpro/utils/root"
	"sync"

	"github.com/spf13/cobra"
)

var (
	localFilePath  string
	remoteFilePath string
	recursive      bool
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "传输本地文件到远程主机",
	Run: func(cmd *cobra.Command, args []string) {
		hostList, err := root.GenHostList(hosts, hostRange, hostNet)
		if err != nil {
			fmt.Println("parse host list failed, err: ", err)
			return
		}

		ch := make(chan bool, goNum)
		wg := new(sync.WaitGroup)

		if recursive {
			for _, host := range hostList {
				c := root.Connection{Host: host, Port: port, Username: username, Password: password, Key: key}
				cp := copy2.Cp{root.Result{Connection: &c}, localFilePath, remoteFilePath}
				fmt.Println(cp.LocalFilePath)
				output, e := copy2.MultiScp(cp, ch, wg)
				if e != nil {
					fmt.Println(e)
				}
				fmt.Println(output)
			}
			return
		}
		for _, host := range hostList {
			wg.Add(1)
			go func(h string) {
				c := root.Connection{Host: h, Port: port, Username: username, Password: password, Key: key}
				cp := copy2.Cp{root.Result{Connection: &c}, localFilePath, remoteFilePath}
				output, e := copy2.Scp(cp, ch, wg)
				if e != nil {
					fmt.Println(e)
					return
				}
				fmt.Println(output)
				return
			}(host)
		}
		wg.Wait()
		close(ch)
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&localFilePath, "src", "s", "", "指定原文件")
	copyCmd.Flags().StringVarP(&remoteFilePath, "dest", "d", "", "指定目标存文件储目录或文件")
	copyCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "递归传输")
}
