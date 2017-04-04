package main

func ExamplePs() {
	runCommand([]string{"arukas", "ps"})
	// Output:
	// CONTAINER ID	IMAGE	COMMAND	CREATED	STATUS	NAME	ENDPOINT
	// d19b004c-0d59-4f4f-955c-5bace7c49a34	nginx:latest		2015-12-21T19:48:17.23+09:00	running	test-con1	test-con1.arukascloud.io
}

func ExamplePsAll() {
	runCommand([]string{"arukas", "ps", "-a"})
	// Output:
	// CONTAINER ID	IMAGE	COMMAND	CREATED	STATUS	NAME	ENDPOINT
	// 2b21fe34-328f-4d7e-8678-726d9eff2b7f	nginx:latest		2015-10-19T15:05:34.843+09:00	interrupted	stopped-container	stopped-container.arukascloud.io
	// d19b004c-0d59-4f4f-955c-5bace7c49a34	nginx:latest		2015-12-21T19:48:17.23+09:00	running	test-con1	test-con1.arukascloud.io
}

func ExamplePsQuiet() {
	*psListAll = false // 前のテストのデータが残っているのでリセットが必要
	runCommand([]string{"arukas", "ps", "-q"})
	// Output:
	// d19b004c-0d59-4f4f-955c-5bace7c49a34
}

func ExamplePsAllQuiet() {
	runCommand([]string{"arukas", "ps", "-a", "-q"})
	// Output:
	// 2b21fe34-328f-4d7e-8678-726d9eff2b7f
	// d19b004c-0d59-4f4f-955c-5bace7c49a34
}

func ExamplePsAllQuietShort() {
	runCommand([]string{"arukas", "ps", "-aq"})
	// Output:
	// 2b21fe34-328f-4d7e-8678-726d9eff2b7f
	// d19b004c-0d59-4f4f-955c-5bace7c49a34
}
