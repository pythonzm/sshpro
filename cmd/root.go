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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	copy2 "sshpro/utils/copy"
	"sshpro/utils/root"
	"sync"
)

var (
	cfgFile   string
	hosts     string
	hostRange string
	hostNet   string
	group     string
	port      string
	username  string
	password  string
	command   string
	//ciphers   []string
	key   string
	goNum int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "sshpro",
	Short:   "ansible的阉割版。。。",
	Version: "1.0.0",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		hostList, err := root.GenHostList(hosts, hostRange, hostNet)
		if err != nil {
			fmt.Println("parse host list failed, err: ", err)
			return
		}

		ch := make(chan bool, goNum)
		wg := new(sync.WaitGroup)

		for _, host := range hostList {
			wg.Add(1)
			go func(h string) {
				c := root.Connection{Host: h, Port: port, Username: username, Password: password, Key: key}
				output, e := c.OutPut(command, wg, ch)
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "指定配置文件 (default is $HOME/.sshpro.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&hosts, "hosts", "", "远程主机IP, 可以是一个或多个，多个用主机用','隔开")
	rootCmd.PersistentFlags().StringVar(&hostRange, "host-range", "", "主机范围，例如：10.1.1.1-10.1.1.254")
	rootCmd.PersistentFlags().StringVar(&hostNet, "host-net", "", "主机段，例如：192.168.1.0/24")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "22", "远程端口")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "root", "远程使用用户")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "P", "", "远程用户密码")
	rootCmd.Flags().StringVarP(&command, "command", "c", "", "执行命令")
	rootCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "主机密钥位置，使用绝对路径")
	rootCmd.PersistentFlags().IntVar(&goNum, "go-num", 5, "并发数")
	rootCmd.PersistentFlags().StringVarP(&group, "group", "g", "", "根据配置文件指定某个组执行命令,多个组用','隔开")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if group != "" {
		if cfgFile != "" {
			// Use config file from the flag.
			viper.SetConfigFile(cfgFile)
		} else {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Search config in home directory with name ".sshpro" (without extension).
			viper.AddConfigPath(home)
			viper.SetConfigName(".sshpro")
		}

		viper.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}

		ch := make(chan bool, goNum)
		wg := new(sync.WaitGroup)
		var (
			conns []root.Connection
			err   error
		)

		if group == "all" {
			conns, err = root.ParseAllGroups()
		} else {
			conns, err = root.ParseGroups(group)
		}
		if err != nil {
			fmt.Println("Can't parse config:", err)
			os.Exit(1)
		}

		if localFilePath != "" && remoteFilePath != "" {

			if recursive {
				for _, c := range conns {
					cp := copy2.Cp{root.Result{Connection: &c}, localFilePath, remoteFilePath}
					output, e := copy2.MultiScp(cp, ch, wg)
					if e != nil {
						fmt.Println(e)
					}
					fmt.Println(output)
				}
				return
			}
			for _, c := range conns {
				wg.Add(1)
				go func(conn root.Connection) {
					cp := copy2.Cp{root.Result{Connection: &conn}, localFilePath, remoteFilePath}
					output, e := copy2.Scp(cp, ch, wg)
					if e != nil {
						fmt.Println(e)
						return
					}
					fmt.Println(output)
					return
				}(c)
			}
			wg.Wait()
			close(ch)
			return
		}

		if command != "" {
			for _, c := range conns {
				wg.Add(1)
				go func(conn root.Connection) {
					output, e := conn.OutPut(command, wg, ch)
					if e != nil {
						fmt.Println(e)
						return
					}
					fmt.Println(output)
					return
				}(c)
			}
			wg.Wait()
			close(ch)
			return
		}

	}
}
