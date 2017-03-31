# Parallel Histogram Example

## Structure

This directory contains a kernel located at `main.go`. It has all the
same commands as the Histogram example.

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
