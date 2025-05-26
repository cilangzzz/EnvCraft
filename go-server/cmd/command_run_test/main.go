package main

import "tsc/pkg/util/server_command"

func main() {
	exector := server_command.New()
	args := []string{"hello world"}
	res, err := exector.ExecuteCommand("echo", args, nil)
	println(res)
	if err != nil {
		panic(err)
	}

}
