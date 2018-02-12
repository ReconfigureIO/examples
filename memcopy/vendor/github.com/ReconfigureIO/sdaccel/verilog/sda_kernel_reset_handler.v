//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// Implementation of SDAccel kernel reset handler. It provides support for
// managing kernel resets under control of the external 'go' and 'done' control
// signals.
//

`timescale 1ns/1ps

module sda_kernel_reset_handler
  (regGoValid, regGoHoldoff, regDoneValid, regDoneStop, kernelGoValid,
  kernelGoHoldoff, kernelDoneValid, kernelDoneStop, sysRstReq, wrapperReset,
  kernelReset, clk);

// Specifies the reset counter size. The kernel reset line will be asserted for
// the time it takes the counter to wrap.
parameter ResetCountSize = 5;

// Specifies the length of the reset pipeline, which allows the synthesis tools
// to build a reset tree if required by using register duplication.
parameter ResetPipeLength = 8;

// Derives the reset counter limit.
parameter ResetCountLimit = (1 << ResetCountSize) - 1;

// Specify the reset controller state space.
parameter [2:0]
  ResetIdle = 0,
  ResetTimeout = 1,
  KernelStarting = 2,
  KernelRunning = 3,
  KernelExited = 4;

// Upstream register interface signals.
input  regGoValid;
output regGoHoldoff;
output regDoneValid;
input  regDoneStop;

// Kernel control go output signals.
output kernelGoValid;
input  kernelGoHoldoff;
input  kernelDoneValid;
output kernelDoneStop;

// Specifies the system reset request signal and generated resets.
input  sysRstReq;
output wrapperReset;
output kernelReset;

// Specifies the clock input. There is no standard synchronous reset.
input clk;

// Reset control state machine signals.
reg [2:0]                resetState_d;
reg [ResetCountSize-1:0] resetCount_d;
reg                      kernelReset_d;
reg                      regGoHoldoff_d;
reg                      regDoneValid_d;
reg                      kernelGoValid_d;
reg                      kernelDoneStop_d;

reg [2:0]                resetState_q;
reg [ResetCountSize-1:0] resetCount_q;
reg                      kernelReset_q;
reg                      regGoHoldoff_q;
reg                      regDoneValid_q;
reg                      kernelGoValid_q;
reg                      kernelDoneStop_q;

// Implements a register with an explicit initialisation value, which will have
// the effect of forcing a reset cycle immediately after loading the FPGA
// netlist. Only works with devices that support bitstream initalisation.
reg resetHandlerEnabled_q = 1'b0;
reg wrapperReset_q;

// Specifies the reset pipeline signals.
reg [ResetPipeLength-1:0] wrapperResetPipe_q;
reg [ResetPipeLength-1:0] kernelResetPipe_q;

// Miscellaneous signals.
integer i;

// Initiate automatic reset on FPGA bitstream load.
always @(posedge clk)
begin
  if (sysRstReq | ~resetHandlerEnabled_q)
  begin
    resetHandlerEnabled_q <= 1'b1;
    wrapperReset_q <= 1'b1;
  end
  else
  begin
    resetHandlerEnabled_q <= 1'b1;
    wrapperReset_q <= 1'b0;
  end
end

// Implement combinatorial logic for reset control state machine.
always @(resetState_q, resetCount_q, kernelReset_q, regGoHoldoff_q, regDoneValid_q,
  kernelGoValid_q, kernelDoneStop_q, regGoValid, regDoneStop, kernelGoHoldoff,
  kernelDoneValid)
begin

  // Hold current state by default.
  resetState_d = resetState_q;
  resetCount_d = resetCount_q;
  kernelReset_d = kernelReset_q;
  regGoHoldoff_d = 1'b1;
  regDoneValid_d = 1'b0;
  kernelGoValid_d = 1'b0;
  kernelDoneStop_d = 1'b1;

  // Implement state machine.
  case (resetState_q)

    // Hold the reset state for the required timeout.
    ResetTimeout :
    begin
      if (resetCount_q == ResetCountLimit [ResetCountSize-1:0])
      begin
        resetState_d = ResetIdle;
      end
      resetCount_d = resetCount_q + 1;
    end

    // Wait for the kernel to accept the go signal.
    KernelStarting :
    begin
      if (kernelGoValid_q & ~kernelGoHoldoff)
      begin
        resetState_d = KernelRunning;
      end
      else
      begin
        kernelGoValid_d = 1'b1;
      end
    end

    // In the kernel runnning state, wait for the 'done' response.
    KernelRunning :
    begin
      if (kernelDoneValid & ~kernelDoneStop_q)
      begin
        resetState_d = KernelExited;
      end
      else
      begin
        kernelDoneStop_d = 1'b0;
      end
    end

    // In the kernel exited state, notify the control registers and place the
    // kernel in reset until the next go request is received.
    KernelExited :
    begin
      if (regDoneValid_q & ~regDoneStop)
      begin
        resetState_d = ResetTimeout;
        kernelReset_d = 1'b1;
      end
      else
      begin
        regDoneValid_d = 1'b1;
      end
    end

    // In the reset idle state, wait for a go request from the register block
    // before releasing the kernel reset.
    ResetIdle :
    begin
      if (regGoValid & ~regGoHoldoff_q)
      begin
        resetState_d = KernelStarting;
        kernelReset_d = 1'b0;
      end
      else
      begin
        regGoHoldoff_d = 1'b0;
      end
    end

    // Treat the unreachable default state as a hard reset. This prevents the
    // Xilinx tools from generating dangling nets if the state encoding is
    // automatically converted to one-hot.
    default:
    begin
      resetState_d = ResetTimeout;
      for (i = 0; i < ResetCountSize; i = i + 1)
        resetCount_d [i] = 1'b0;
      kernelReset_d = 1'b1;
      regGoHoldoff_d = 1'b1;
      regDoneValid_d = 1'b0;
      kernelGoValid_d = 1'b0;
      kernelDoneStop_d = 1'b1;
    end
  endcase

end

// Implement sequential logic for reset control state machine.
always @(posedge clk)
begin
  if (wrapperReset_q)
  begin
    resetState_q <= ResetTimeout;
    for (i = 0; i < ResetCountSize; i = i + 1)
      resetCount_q [i] <= 1'b0;
    kernelReset_q <= 1'b1;
    regGoHoldoff_q <= 1'b1;
    regDoneValid_q <= 1'b0;
    kernelGoValid_q <= 1'b0;
    kernelDoneStop_q <= 1'b1;
  end
  else
  begin
    resetState_q <= resetState_d;
    resetCount_q <= resetCount_d;
    kernelReset_q <= kernelReset_d;
    regGoHoldoff_q <= regGoHoldoff_d;
    regDoneValid_q <= regDoneValid_d;
    kernelGoValid_q <= kernelGoValid_d;
    kernelDoneStop_q <= kernelDoneStop_d;
  end
end

assign regGoHoldoff = regGoHoldoff_q;
assign regDoneValid = regDoneValid_q;
assign kernelGoValid = kernelGoValid_q;
assign kernelDoneStop = kernelDoneStop_q;

// Implement reset output pipelines.
always @(posedge clk)
begin
  if (wrapperReset_q)
    for (i = 0; i < ResetPipeLength; i = i + 1)
       wrapperResetPipe_q [i] <= 1'b1;
  else
    wrapperResetPipe_q <= { 1'b0, wrapperResetPipe_q [ResetPipeLength-1:1] };
end

always @(posedge clk)
begin
  if (kernelReset_q)
    for (i = 0; i < ResetPipeLength; i = i + 1)
       kernelResetPipe_q [i] <= 1'b1;
  else
    kernelResetPipe_q <= { 1'b0, kernelResetPipe_q [ResetPipeLength-1:1] };
end

assign wrapperReset = wrapperResetPipe_q [0];
assign kernelReset = kernelResetPipe_q [0];

endmodule
