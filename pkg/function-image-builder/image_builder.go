/*
Copyright (c) 2016-2017 Bitnami

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

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	lbuilder "github.com/kubeless/kubeless/pkg/function-image-builder/layer-builder"
	"github.com/spf13/cobra"
)

var globalUsage = `` //TODO: add explanation

func init() {
	layerCmd.Flags().Bool("insecure", false, "Disable TLS verification.")
	layerCmd.Flags().StringP("src", "", "", "Source image reference. F.e. dir://path/to/image")
	layerCmd.Flags().StringP("src-creds", "", "", "Source image credentials in case it is a private registry. F.e. user:my_pass")
	layerCmd.Flags().StringP("dst", "", "", "Destination image reference. F.e. docker://user/image")
	layerCmd.Flags().StringP("dst-creds", "", "", "Destination credentials in case it is a docker registry. F.e. user:my_pass")
	layerCmd.Flags().StringP("cwd", "", "", "Working directory")
}

func runCommand(command string, args []string) error {
	cmd := exec.Command(command, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scannerStdout := bufio.NewScanner(stdout)
	scannerStdout.Split(bufio.ScanLines)
	for scannerStdout.Scan() {
		m := scannerStdout.Text()
		fmt.Fprintln(os.Stdout, m)
	}
	scannerStderr := bufio.NewScanner(stderr)
	scannerStderr.Split(bufio.ScanLines)
	for scannerStderr.Scan() {
		m := scannerStderr.Text()
		fmt.Fprintln(os.Stderr, m)
	}

	return cmd.Wait()
}

func skopeoCopy(src, dst, srcCreds, dstCreds string, insecure bool) error {
	command := "skopeo"
	args := []string{"copy"}
	if srcCreds != "" {
		args = append(args, "--src-creds", srcCreds)
	}
	if dstCreds != "" {
		args = append(args, "--dest-creds", dstCreds)
	}
	if insecure {
		args = append(args, "--src-tls-verify=false", "--dest-tls-verify=false")
	}
	args = append(args, src, dst)
	return runCommand(command, args)
}

var layerCmd = &cobra.Command{
	Use:   "add-layer <tar> FLAG",
	Short: "Add tar as a image layer",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Need exactly one argument - layer tar")
		}

		layerTar := args[0]

		srcImage, err := cmd.Flags().GetString("src")
		if err != nil {
			log.Fatal(err)
		}
		if srcImage == "" {
			log.Fatal("Need specify the source image using the flag --src")
		}

		dstImage, err := cmd.Flags().GetString("dst")
		if err != nil {
			log.Fatal(err)
		}
		if dstImage == "" {
			log.Fatal("Need specify the destination image using the flag --dst")
		}

		srcCreds, err := cmd.Flags().GetString("src-creds")
		if err != nil {
			log.Fatal(err)
		}

		dstCreds, err := cmd.Flags().GetString("dst-creds")
		if err != nil {
			log.Fatal(err)
		}

		workDir, err := cmd.Flags().GetString("cwd")
		if err != nil {
			log.Fatal(err)
		}
		if workDir == "" {
			workDir, err = ioutil.TempDir("", "build")
			if err != nil {
				log.Fatal(err)
			}
		}

		insecure, err := cmd.Flags().GetBool("insecure")
		if err != nil {
			log.Fatal(err)
		}

		// Store src image
		err = skopeoCopy(srcImage, fmt.Sprintf("dir://%s", workDir), srcCreds, dstCreds, insecure)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Succesfully stored base image ", srcImage, " at ", workDir)

		// Add layer
		err = lbuilder.AddTarToLayer(workDir, layerTar)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Added layer ", layerTar, " in ", workDir)

		// Publish new image
		err = skopeoCopy(fmt.Sprintf("dir://%s", workDir), dstImage, srcCreds, dstCreds, insecure)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Succesfully stored final image at ", dstImage)
	},
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "imbuilder",
		Short: "Pulls an image and push a new one including a tar file as a new layer",
		Long:  globalUsage,
	}

	cmd.AddCommand(layerCmd)
	return cmd
}

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
