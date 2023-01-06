//go:build linux
// +build linux

/*
Copyright © 2021 SUSE LLC

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
	"errors"
	"fmt"
	"os"
	"path"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

var kubeconfigViper = viper.New()

// kubeconfigCmd represents the kubeconfig command, used to set up a symlink on
// the Linux side to point at the Windows-side kubeconfig.  Note that we must
// pass the kubeconfig path in as an environment variable to take advantage of
// the path translation capabilities of WSL2 interop.
var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Set up ~/.kube/config in the WSL2 environment",
	Long:  `This command configures the Kubernetes configuration inside a WSL2 distribution.`,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := kubeconfigViper.GetString("kubeconfig")
		enable := kubeconfigViper.GetBool("enable")
		show := kubeconfigViper.GetBool("show")

		if configPath == "" {
			return errors.New("Windows kubeconfig not supplied")
		}

		_, err := os.Stat(configPath)
		if err != nil {
			return fmt.Errorf("could not open Windows kubeconfig: %w", err)
		}
		cmd.SilenceUsage = true

		configDir := path.Join(homedir.HomeDir(), ".kube")
		linkPath := path.Join(configDir, "config")
		if show {
			// The output is "true", "false", or an error message for UI.
			// We will only return nil in this path.
			fmt.Println(getLinkStatus(configDir, linkPath, configPath))
			return nil
		}
		if enable {
			return createLinkToLocalConfig(configDir, linkPath, configPath)
		} else {
			return deleteLinkToLocalConfig(configDir, linkPath, configPath)
			if errors.Is(err, syscall.EINVAL) {
				// See if the source's parent dir is a link to the target's parent dir
				target, err := os.Readlink(configDir)
				if err != nil {
					return nil
				}
				if target == path.Dir(configPath) {
					err = os.Remove(linkPath)
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						return err
					}
				}
			}
		}
		return nil
	},
}

func init() {
	kubeconfigCmd.PersistentFlags().Bool("enable", true, "Set up config file")
	kubeconfigCmd.PersistentFlags().String("kubeconfig", "", "Path to Windows kubeconfig, in /mnt/... form.")
	kubeconfigCmd.PersistentFlags().Bool("show", false, "Get the current state rather than set it")
	kubeconfigViper.AutomaticEnv()
	kubeconfigViper.BindPFlags(kubeconfigCmd.PersistentFlags())
	rootCmd.AddCommand(kubeconfigCmd)
}

// configDir = path.Parent(linkPath)
func getLinkStatus(configDir string, linkPath string, configPath string) string {
	target, err := os.Readlink(linkPath)
	if err == nil {
		// For a symlink pointing elsewhere, we assume we can overwrite.
		return fmt.Sprintf("%v", target == configPath)
	}
	if errors.Is(err, os.ErrNotExist) {
		return "false"
	}
	if errors.Is(err, syscall.EINVAL) {
		// See if the source's parent dir is a link to the target's parent dir
		target, err := os.Readlink(configDir)
		if err != nil {
			return fmt.Sprintf("File %s exists and is not a symlink", linkPath)
		}
		// If ~/.kube => the local WSL /DRIVE/mnt/Users/USER/.kube, it's good
		return fmt.Sprintf("%v", target == path.Dir(configPath))
	}
	return err.Error()
}

func createLinkToLocalConfig(configDir string, linkPath string, configPath string) error {
	err = os.Mkdir(configDir, 0o750)
	if err != nil && !errors.Is(err, os.ErrExist) {
		// The error already contains the full path, we can't do better.
		return err
	}
	err = os.Symlink(configPath, linkPath)
	if err != nil && errors.Is(err, os.ErrExist) {
		// If it already exists, do nothing; even if it's not a symlink.
		return nil
	}
	return err
}

func deleteLinkToLocalConfig(configDir string, linkPath string, configPath string) error {
	// No need to create if we want to remove it
	target, err := os.Readlink(linkPath)
	if err == nil {
		if target == configPath {
			err = os.Remove(linkPath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if !errors.Is(err, syscall.EINVAL) {
		return err
	}
	// If ~/.kube is a link to the mounted /mnt/DRIVE/Users/USER/.kube, delete the link also.
	target, err = os.Readlink(configDir)
	if err != nil {
		return nil
	}
	if target == path.Dir(configPath) {
		err = os.Remove(linkPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
