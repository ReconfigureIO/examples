# Histogram Example

## Structure

This directory contains a kernel located at `main.go`. It also has a
command, `test-histogram` located at `cmd/test-histogram/main.go`.

## Testing

To run this example in a simulator, execute the following:

```
reco test test-histogram
```

This will simulate the kernel using a hardware simulator, and test it
using the `test-histogram` command.

## Building

```
reco build
```

This will build your commands and kernel for execution on hardware.
