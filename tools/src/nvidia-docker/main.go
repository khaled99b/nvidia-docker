// Copyright (c) 2015-2016, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"log"
	"net/url"
	"os"
	"runtime/debug"

	"github.com/NVIDIA/nvidia-docker/tools/src/docker"
	"github.com/NVIDIA/nvidia-docker/tools/src/nvidia"
)

var (
	Host *url.URL
	GPU  []string
)

func init() {
	log.SetPrefix(os.Args[0] + " | ")
	LoadEnvironment()
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func exit() {
	code := 0
	if r := recover(); r != nil {
		code = 1
		log.Printf("Error: %v", r)
		if os.Getenv("NV_DEBUG") != "" {
			debug.PrintStack()
		}
	}
	os.Exit(code)
}

func GenerateDockerArgs(image string) []string {
	vols, err := VolumesNeeded(image)
	assert(err)
	if vols == nil {
		return nil
	}
	if Host != nil {
		args, err := GenerateRemoteArgs(image, vols)
		assert(err)
		return args
	}
	args, err := GenerateLocalArgs(image, vols)
	assert(err)
	return args
}

func main() {
	var option string
	var n int

	args := os.Args[1:]
	defer exit()

	assert(nvidia.Init())
	defer func() { assert(nvidia.Shutdown()) }()

	command, i, err := docker.ParseArgs(args)
	assert(err)
	if command == "create" || command == "run" || command == "volume" {
		option, n, err = docker.ParseArgs(args[i+1:], command)
		i += n + 1
		assert(err)
	}
	switch command {
	case "create":
		fallthrough
	case "run":
		if option != "" {
			a := GenerateDockerArgs(option)
			args = append(args[:i], append(a, args[i:]...)...)
		}
	case "volume":
		if option == "setup" {
			assert(CreateLocalVolumes())
			return
		}
	default:
	}

	assert(nvidia.LoadUVM())
	assert(docker.Docker(args...))
}
