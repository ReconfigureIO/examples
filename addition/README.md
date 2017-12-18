# Addition Example

## Structure

This directory contains a code for and FPGA located at `main.go`. It also has a
command, `test-addition` located at `cmd/test-addition/main.go`.

## Testing

To run this example in a simulator, execute the following:

```
reco test test-addition
```

This will simulate the FPGA code using a hardware simulator, and test it
using the `test-addition` command.

## Building

```
reco build
```

This will build your commands and FPGA code for execution on hardware.
