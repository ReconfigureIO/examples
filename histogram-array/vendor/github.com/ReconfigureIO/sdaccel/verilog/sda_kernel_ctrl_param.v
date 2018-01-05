//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// Implements the parameter RAM block which is provided on the AXI control bus
// for assigning kernel parameters prior to running kernel code.
//

`timescale 1ns/1ps

module sda_kernel_ctrl_param
  (regReq, regAck, regWriteEn, regAddr, regWData, regWStrb, regRData,
  paramAddrValid, paramAddr, paramAddrStop, paramDataValid, paramData,
  paramDataStop, clk, srst);

// Specifies the width of the register address bus.
parameter RegAddrWidth = 12;

// Specifies the base address of the parameter block.
// The default is to reserve space for 16 32-bit Verilog wrapper registers.
parameter ParamAddrBase = 64;

// Specifies the upper address of the parameter block.
parameter ParamAddrTop = 4095;

// Slave side simple register interface signals. Note that all outputs are
// driven low when inactive so that they can be ORed together with other
// register block implementations.
input                    regReq;
output                   regAck;
input                    regWriteEn;
input [RegAddrWidth-1:0] regAddr;
input [31:0]             regWData;
input [3:0]              regWStrb;
output [31:0]            regRData;

// Kernel interface parameter access signals.
input         paramAddrValid;
input [31:0]  paramAddr;
output        paramAddrStop;
output        paramDataValid;
output [31:0] paramData;
input         paramDataStop;

// System level signals.
input clk;
input srst;

// Inferred RAM.
reg [31:0] ramArray [(ParamAddrTop-ParamAddrBase+1)/4-1:0];

// Pipelined register interface input inputs.
reg                    regReq_q;
reg                    regReadReq_q;
reg                    regWriteReq_q;
reg [31:0]             regWData_q;
reg [3:0]              regWStrb_q;
reg [RegAddrWidth-3:0] regAddr_q;

// Pipelined register interface output inputs.
reg                    regAck_q;
reg                    regReadDone_q;
reg                    regReadValid_q;
reg [31:0]             regRData_q;
reg [31:0]             regReadData_q;
reg [31:0]             regPipeData_q;

// Pipelined parameter RAM access signals.
reg        paramAddrValid_q;
reg [31:0] paramAddr_q;
reg        paramDataValid_q;
reg [31:0] paramData_q;

// Parameter RAM access backpressure signals.
wire        pmAddrStop;
wire        pmReadStop;
wire        pmPipeStop;

// Parameter RAM access pipeline.
reg [RegAddrWidth-3:0] pmAddr_q;
reg [1:0]              pmAddrAlign_q;
reg                    pmAddrValid_q;
reg [31:0]             pmReadData_q;
reg [1:0]              pmReadAlign_q;
reg                    pmReadValid_q;
reg [31:0]             pmPipeData_q;
reg [1:0]              pmPipeAlign_q;
reg                    pmPipeValid_q;
reg [31:0]             pmDataAligned;

// Miscellaneous signals.
wire [RegAddrWidth-1:0] regParamAddrBase = ParamAddrBase [RegAddrWidth-1:0];
wire [RegAddrWidth-1:0] regParamAddrTop = ParamAddrTop [RegAddrWidth-1:0];
integer i;

// Implement pipelined register input signals. Assumes that there are no back
// to back transactions, so we can use rising edge detection on the request line.
// verilator lint_off CMPCONST
always @(posedge clk)
begin
  if (srst)
  begin
    regReq_q <= 1'b0;
    regReadReq_q <= 1'b0;
    regWriteReq_q <= 1'b0;
    regWData_q <= 32'b0;
    regWStrb_q <= 4'b0;
    for (i = 0; i < RegAddrWidth-2; i = i + 1)
      regAddr_q[i] <= 1'b0;
  end
  else
  begin
    regReq_q <= regReq;
    regWData_q <= regWData;
    regWStrb_q <= regWStrb;
    if ((regAddr < regParamAddrBase) || (regAddr > regParamAddrTop))
    begin
      regReadReq_q <= 1'b0;
      regWriteReq_q <= 1'b0;
      for (i = 0; i < RegAddrWidth-2; i = i + 1)
        regAddr_q[i] <= 1'b0;
    end
    else
    begin
      regReadReq_q <= regReq & ~regReq_q & ~regWriteEn;
      regWriteReq_q <= regReq & ~regReq_q & regWriteEn;
      regAddr_q <= regAddr[RegAddrWidth-1:2] - (ParamAddrBase/4);
    end
  end
end
// verilator lint_on CMPCONST

// Implement pipelined register output signals.
always @(posedge clk)
begin
  if (srst)
  begin
    regAck_q <= 1'b0;
    regReadDone_q <= 1'b0;
    regReadValid_q <= 1'b0;
    regRData_q <= 32'b0;
  end
  else
  begin
    regAck_q <= regReadValid_q | regWriteReq_q;
    regReadDone_q <= regReadReq_q;
    regReadValid_q <= regReadDone_q;
    regRData_q <= regReadValid_q ? regPipeData_q : 32'b0;
  end
end

assign regAck = regAck_q;
assign regRData = regRData_q;

// Implement pipelined parameter address inputs.
always @(posedge clk)
begin
  if (srst)
  begin
    paramAddrValid_q <= 1'b0;
    paramAddr_q <= 32'b0;
  end
  else if (paramAddrValid_q)
  begin
    paramAddrValid_q <= pmAddrStop;
  end
  else
  begin
    paramAddrValid_q <= paramAddrValid;
    paramAddr_q <= paramAddr;
  end
end

assign paramAddrStop = paramAddrValid_q;

// Implement the parameter data RAM access backpressure signals.
assign pmAddrStop = pmReadStop & pmAddrValid_q;
assign pmReadStop = pmPipeStop & pmReadValid_q;
assign pmPipeStop = paramDataValid_q & pmPipeValid_q;

// Implement parameter access input pipeline.
always @(posedge clk)
begin
  if (srst)
  begin
    for (i = 0; i < RegAddrWidth-2; i = i + 1)
      pmAddr_q[i] <= 1'b0;
    pmAddrValid_q <= 1'b0;
    pmReadValid_q <= 1'b0;
    pmPipeValid_q <= 1'b0;
  end
  else
  begin
    if (~pmAddrStop)
    begin
      pmAddrValid_q <= paramAddrValid_q;
      if ((paramAddr_q < ParamAddrBase) || (paramAddr_q > ParamAddrTop))
      begin
        for (i = 0; i < RegAddrWidth-2; i = i + 1)
          pmAddr_q[i] <= 1'b0;
        pmAddrAlign_q <= 2'b0;
      end
      else
      begin
        pmAddr_q <= paramAddr_q[RegAddrWidth-1:2] - (ParamAddrBase/4);
        pmAddrAlign_q <= paramAddr_q[1:0];
      end
    end
    if (~pmReadStop)
    begin
      pmReadValid_q <= pmAddrValid_q;
      pmReadAlign_q <= pmAddrAlign_q;
    end
    if (~pmPipeStop)
    begin
      pmPipeValid_q <= pmReadValid_q;
      pmPipeAlign_q <= pmReadAlign_q;
    end
  end
end

// Perform data alignment on read data. Uses the least significant bits of the
// parameter address to rotate the addressed byte into the LSB position. When
// combined with a suitable type cast in the kernel code, this allows byte and
// half word parameter values to be addressed on byte and half word boundaries.
always @(pmPipeAlign_q, pmPipeData_q)
begin
  case (pmPipeAlign_q)
    2'b11 : pmDataAligned = {pmPipeData_q [23:0], pmPipeData_q [31:24]};
    2'b10 : pmDataAligned = {pmPipeData_q [15:0], pmPipeData_q [31:16]};
    2'b01 : pmDataAligned = {pmPipeData_q [7:0], pmPipeData_q [31:8]};
    default: pmDataAligned = pmPipeData_q;
  endcase
end

// Provide output pipeline register for read data.
always @(posedge clk)
begin
  if (srst)
  begin
    paramDataValid_q <= 1'b0;
    paramData_q <= 32'b0;
  end
  else if (paramDataValid_q)
  begin
    paramDataValid_q <= paramDataStop;
  end
  else
  begin
    paramDataValid_q <= pmPipeValid_q;
    paramData_q <= pmDataAligned;
  end
end

assign paramDataValid = paramDataValid_q;
assign paramData = paramData_q;

// Implement parameter RAM.
always @(posedge clk)
begin

  // SELF parameter pipeline is gated for backpressure.
  if (~pmReadStop)
  begin
    pmReadData_q <= ramArray [pmAddr_q];
  end

  // Register read pipeline is a single cycle delay.
  regReadData_q <= ramArray [regAddr_q];

  // Apply write enable strobes.
  if (regWriteReq_q)
  begin
    if (regWStrb_q[0])
      ramArray [regAddr_q][7:0] <= regWData_q [7:0];
    if (regWStrb_q[1])
      ramArray [regAddr_q][15:8] <= regWData_q [15:8];
    if (regWStrb_q[2])
      ramArray [regAddr_q][23:16] <= regWData_q [23:16];
    if (regWStrb_q[3])
      ramArray [regAddr_q][31:24] <= regWData_q [31:24];
  end
end

// Pipeline read data for improved timing.
always @(posedge clk)
begin

  // SELF parameter pipeline is gated for backpressure.
  if (~pmPipeStop)
  begin
    pmPipeData_q <= pmReadData_q;
  end

  // Register read pipeline is a single cycle delay.
  regPipeData_q <= regReadData_q;

end

endmodule
