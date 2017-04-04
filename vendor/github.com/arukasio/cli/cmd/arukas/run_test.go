package main

func ExampleRun() {
	runCommand([]string{"arukas", "run", "nginx:latest", "--instances", "1", "--mem", "256", "-p", "80:tcp"})
	// Output:
	// ID	IMAGE	CREATED	STATUS	NAME	ENDPOINT
	// 2b21fe34-328f-4d7e-8678-726d9eff2b7f	nginx:latest	2015-10-19T15:05:34.843+09:00	interrupted	stopped-container	stopped-container.arukascloud.io
}

func ExampleRunWithWrongPortProtocol() {
	runCommand([]string{"arukas", "run", "nginx:latest", "--instances", "1", "--mem", "256", "-p", "80:80"})
	// Output:
	// Port protocol must be "tcp" or "udp"
}
