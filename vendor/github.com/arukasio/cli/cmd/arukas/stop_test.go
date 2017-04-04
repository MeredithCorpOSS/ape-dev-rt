package main

func ExampleStop() {
	runCommand([]string{"arukas", "stop", "d19b004c-0d59-4f4f-955c-5bace7c49a34"})
	// Output:
	// Stopping...
}

// func TestStopAlreadyStopped(t *testing.T) {
// 	runCommand([]string{"arukas", "stop", "2b21fe34-328f-4d7e-8678-726d9eff2b7f"})
// 	if ExitCode != 1 {
// 		t.Errorf(("ExitCode got: %d, want: 1"), ExitCode)
// 	}
// }

// func ExampleStopAlreadyStopped() {
// 	runCommand([]string{"arukas", "stop", "2b21fe34-328f-4d7e-8678-726d9eff2b7f"})
// 	// Output:
// 	// Failed to stop the container
// }
