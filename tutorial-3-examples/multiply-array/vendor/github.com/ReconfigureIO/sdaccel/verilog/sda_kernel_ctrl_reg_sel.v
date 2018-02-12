//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// Implementation of SDAccel kernel entity control register selection unit. It
// maps the specified number of AXI slave interface registers at the start of
// the AXI address space to simple wrapper control registers and then maps the
// remaining locations to the AXI interface handler in the generated code.
//

`timescale 1ns/1ps

module sda_kernel_ctrl_reg_sel
  (sAxiAWValid, sAxiAWReady, sAxiAWAddr, sAxiWValid, sAxiWReady, sAxiWData,
  sAxiWStrb, sAxiBValid, sAxiBReady, sAxiBResp, sAxiARValid, sAxiARReady,
  sAxiARAddr, sAxiRValid, sAxiRReady, sAxiRData, sAxiRResp, mAxiAWValid,
  mAxiAWReady, mAxiAWAddr, mAxiWValid, mAxiWReady, mAxiWData, mAxiWStrb,
  mAxiBValid, mAxiBReady, mAxiBResp, mAxiARValid, mAxiARReady, mAxiARAddr,
  mAxiRValid, mAxiRReady, mAxiRData, mAxiRResp, regReq, regAck, regWriteEn,
  regAddr, regWData, regWStrb, regRData, clk, srst);

// Specifies the width of the AXI address bus.
parameter AddrWidth = 16;

// Specifies the width of the local register set address bus.
parameter RegAddrWidth = 8;

// Specify the upper address location which is mapped to the local register set.
parameter RegAddrTop = 255;

// Slave side AXI write address channel signals.
input                 sAxiAWValid;
output                sAxiAWReady;
input [AddrWidth-1:0] sAxiAWAddr;

// Slave side AXI write data channel signals.
input        sAxiWValid;
output       sAxiWReady;
input [31:0] sAxiWData;
input [3:0]  sAxiWStrb;

// Slave side AXI write acknowledgement channel signals.
output       sAxiBValid;
input        sAxiBReady;
output [1:0] sAxiBResp;

// Slave side AXI read address channel signals.
input                 sAxiARValid;
output                sAxiARReady;
input [AddrWidth-1:0] sAxiARAddr;

// Slave side AXI read data channel signals.
output        sAxiRValid;
input         sAxiRReady;
output [31:0] sAxiRData;
output [1:0]  sAxiRResp;

// Master side AXI write address channel signals.
output                 mAxiAWValid;
input                  mAxiAWReady;
output [AddrWidth-1:0] mAxiAWAddr;

// Master side AXI write data channel signals.
output        mAxiWValid;
input         mAxiWReady;
output [31:0] mAxiWData;
output [3:0]  mAxiWStrb;

// Master side AXI write acknowledgement channel signals.
input       mAxiBValid;
output      mAxiBReady;
input [1:0] mAxiBResp;

// Master side AXI read address channel signals.
output                 mAxiARValid;
input                  mAxiARReady;
output [AddrWidth-1:0] mAxiARAddr;

// Slave side AXI read data channel signals.
input        mAxiRValid;
output       mAxiRReady;
input [31:0] mAxiRData;
input [1:0]  mAxiRResp;

// Master side simple register interface signals.
output                    regReq;
input                     regAck;
output                    regWriteEn;
output [RegAddrWidth-1:0] regAddr;
output [31:0]             regWData;
output [3:0]              regWStrb;
input  [31:0]             regRData;

// System level signals.
input          clk;
input          srst;

// AXI write address channel register signals.
wire                 sAxiAWPending;
reg                  sAxiAWClear;
wire [AddrWidth-1:0] sAxiAWAddrReg;
reg                  mAxiAWPush;
wire                 mAxiAWBlocked;
wire [AddrWidth-1:0] mAxiAWAddrReg;

// AXI write data channel register signals.
wire        sAxiWPending;
reg         sAxiWClear;
wire [31:0] sAxiWDataReg;
wire [3:0]  sAxiWStrbReg;
reg         mAxiWPush;
wire        mAxiWBlocked;
wire [31:0] mAxiWDataReg;
wire [3:0]  mAxiWStrbReg;

// AXI read address channel register signals.
wire                 sAxiARPending;
reg                  sAxiARClear;
wire [AddrWidth-1:0] sAxiARAddrReg;
reg                  mAxiARPush;
wire                 mAxiARBlocked;
wire [AddrWidth-1:0] mAxiARAddrReg;

// AXI write response channel register signals.
wire       mAxiBPending;
reg        mAxiBClear;
wire [1:0] mAxiBRespReg;
reg        sAxiBPush;
wire       sAxiBBlocked;
reg  [1:0] sAxiBRespReg;

// AXI read response channel register signals.
wire        mAxiRPending;
reg         mAxiRClear;
wire [31:0] mAxiRDataReg;
wire [1:0]  mAxiRRespReg;
reg         sAxiRPush;
wire        sAxiRBlocked;
reg  [31:0] sAxiRDataReg;
reg  [1:0]  sAxiRRespReg;

// Specify the state space used to select the AXI transaction mode.
parameter [3:0]
  Idle = 0,
  RegReadStart = 1,
  RegReadActive = 2,
  RegWriteStart = 3,
  RegWriteActive = 4,
  AxiReadStart = 5,
  AxiReadActive = 6,
  AxiWriteStart = 7,
  AxiWriteData = 8,
  AxiWriteActive = 9;

// Specify AXI state machine registers.
reg [3:0]              axiState_d;
reg                    regReq_d;
reg                    regWriteEn_d;
reg [RegAddrWidth-1:0] regAddr_d;
reg [31:0]             regWData_d;
reg [3:0]              regWStrb_d;

reg [3:0]              axiState_q;
reg                    regReq_q;
reg                    regWriteEn_q;
reg [RegAddrWidth-1:0] regAddr_q;
reg [31:0]             regWData_q;
reg [3:0]              regWStrb_q;

// Miscellaneous signals.
wire [AddrWidth-1:0] regAddrTop = RegAddrTop [AddrWidth-1:0];
integer i;

// Instantiate input registers for slave side AXI write address channel.
sda_kernel_ctrl_reg_sel_axi_inreg_x1 #(AddrWidth) sAxiAWReg_u
  (sAxiAWValid, sAxiAWReady, sAxiAWAddr, sAxiAWPending, sAxiAWClear,
  sAxiAWAddrReg, clk, srst);

// Instantiate input registers for slave side AXI data channel.
sda_kernel_ctrl_reg_sel_axi_inreg_x2 #(32, 4) sAxiWReg_u
  (sAxiWValid, sAxiWReady, sAxiWData, sAxiWStrb, sAxiWPending,
  sAxiWClear, sAxiWDataReg, sAxiWStrbReg, clk, srst);

// Instantiate input registers for slave side AXI read address channel.
sda_kernel_ctrl_reg_sel_axi_inreg_x1 #(AddrWidth) sAxiARReg_u
  (sAxiARValid, sAxiARReady, sAxiARAddr, sAxiARPending, sAxiARClear,
  sAxiARAddrReg, clk, srst);

// Instantiate input register for master side AXI write acknowledgement.
sda_kernel_ctrl_reg_sel_axi_inreg_x1 #(2) mAxiBReg_u
  (mAxiBValid, mAxiBReady, mAxiBResp, mAxiBPending, mAxiBClear, mAxiBRespReg,
  clk, srst);

// Instantiate input register for master side AXI read data channel.
sda_kernel_ctrl_reg_sel_axi_inreg_x2 #(32, 2) mAxiRReg_u
  (mAxiRValid, mAxiRReady, mAxiRData, mAxiRResp, mAxiRPending,
  mAxiRClear, mAxiRDataReg, mAxiRRespReg, clk, srst);

// Instantate output register for master side AXI write address channel.
sda_kernel_ctrl_reg_sel_axi_outreg_x1 #(AddrWidth) mAxiAWReg_u
  (mAxiAWPush, mAxiAWBlocked, mAxiAWAddrReg, mAxiAWValid, mAxiAWReady,
  mAxiAWAddr, clk, srst);

// Instantiate output register for master side AXI write data channel.
sda_kernel_ctrl_reg_sel_axi_outreg_x2 #(32, 4) mAxiWReg_u
  (mAxiWPush, mAxiWBlocked, mAxiWDataReg, mAxiWStrbReg, mAxiWValid,
  mAxiWReady, mAxiWData, mAxiWStrb, clk, srst);

// Instantiate output register for master side AXI read address channel.
sda_kernel_ctrl_reg_sel_axi_outreg_x1 #(AddrWidth) mAxiARReg_u
  (mAxiARPush, mAxiARBlocked, mAxiARAddrReg, mAxiARValid, mAxiARReady,
  mAxiARAddr, clk, srst);

// Instantiate output register for slave side AXI write acknowledgement.
sda_kernel_ctrl_reg_sel_axi_outreg_x1 #(2) sAxiBReg_u
  (sAxiBPush, sAxiBBlocked, sAxiBRespReg, sAxiBValid, sAxiBReady, sAxiBResp,
  clk, srst);

// Instantiate output register for slave side AXI read data channel.
sda_kernel_ctrl_reg_sel_axi_outreg_x2 #(32, 2) sAxiRReg_u
  (sAxiRPush, sAxiRBlocked, sAxiRDataReg, sAxiRRespReg, sAxiRValid,
  sAxiRReady, sAxiRData, sAxiRResp, clk, srst);

// Pass through AXI signals where possible.
assign mAxiAWAddrReg = sAxiAWAddrReg;
assign mAxiWDataReg = sAxiWDataReg;
assign mAxiWStrbReg = sAxiWStrbReg;
assign mAxiARAddrReg = sAxiARAddrReg;

// Implement combinatorial logic for selecting AXI transaction mode.
always @(axiState_q, regReq_q, regWriteEn_q, regAddr_q, regWData_q, regWStrb_q,
  sAxiAWPending, sAxiAWAddrReg, sAxiWPending, sAxiWDataReg, sAxiWStrbReg,
  sAxiBBlocked, sAxiARPending, sAxiARAddrReg, sAxiRBlocked, mAxiRPending,
  mAxiRDataReg, mAxiRRespReg, mAxiAWBlocked, mAxiWBlocked, mAxiBPending,
  mAxiARBlocked, mAxiBRespReg, regAck, regRData, regAddrTop)
begin

  // Preserve current state by default.
  axiState_d = axiState_q;
  regReq_d = regReq_q;
  regWriteEn_d = regWriteEn_q;
  regAddr_d = regAddr_q;
  regWData_d = regWData_q;
  regWStrb_d = regWStrb_q;

  // Set default read assignment to register inputs with AXI 'OKAY' response.
  sAxiRPush = 1'b0;
  sAxiRDataReg = regRData;
  sAxiRRespReg = 2'b0;

  // Set default write status assigment to AXI 'OKAY' response.
  sAxiBPush = 1'b0;
  sAxiBRespReg = 2'b0;

  // Disable AXI register clear strobes by default.
  sAxiAWClear = 1'b0;
  sAxiWClear = 1'b0;
  sAxiARClear = 1'b0;
  mAxiBClear = 1'b0;
  mAxiRClear = 1'b0;

  // Disable AXI master push strobes by default.
  mAxiAWPush = 1'b0;
  mAxiWPush = 1'b0;
  mAxiARPush = 1'b0;

  // Implement state machine.
  case (axiState_q)

    // In the idle state, wait until the AXI write or read address inputs are
    // ready. Writes are prioritised over reads.
    // verilator lint_off CMPCONST
    Idle :
    begin
      if (sAxiAWPending)
      begin
        if (sAxiAWAddrReg <= regAddrTop)
          axiState_d = RegWriteStart;
        else
          axiState_d = AxiWriteStart;
      end
      else if (sAxiARPending)
      begin
        if (sAxiARAddrReg <= regAddrTop)
          axiState_d = RegReadStart;
        else
          axiState_d = AxiReadStart;
      end
    end
    // verilator lint_on CMPCONST

    // Initiate read transactions on the local register interface.
    RegReadStart :
    begin
      if (~sAxiRBlocked)
      begin
        axiState_d = RegReadActive;
        regReq_d = 1'b1;
        regWriteEn_d = 1'b0;
        regAddr_d = sAxiARAddrReg [RegAddrWidth-1:0];
      end
    end

    // Process active read requests.
    RegReadActive :
    begin
      if (regAck)
      begin
        axiState_d = Idle;
        regReq_d = 1'b0;
        regWriteEn_d = 1'b0;
        sAxiRPush = 1'b1;
        sAxiARClear = 1'b1;
      end
    end

    // Initiate write transactions to the local register interface.
    RegWriteStart :
    begin
      if (sAxiWPending & ~sAxiBBlocked)
      begin
        axiState_d = RegWriteActive;
        regReq_d = 1'b1;
        regWriteEn_d = 1'b1;
        regAddr_d = sAxiAWAddrReg [RegAddrWidth-1:0];
        regWData_d = sAxiWDataReg;
        regWStrb_d = sAxiWStrbReg;
      end
    end

    // Process active write requests.
    RegWriteActive :
    begin
      if (regAck)
      begin
        axiState_d = Idle;
        regReq_d = 1'b0;
        regWriteEn_d = 1'b0;
        sAxiBPush = 1'b1;
        sAxiAWClear = 1'b1;
        sAxiWClear = 1'b1;
      end
    end

    // Initiate read transaction on the AXI master side.
    AxiReadStart :
    begin
      if (~mAxiARBlocked)
      begin
        axiState_d = AxiReadActive;
        mAxiARPush = 1'b1;
        sAxiARClear = 1'b1;
      end
    end

    // Complete read transaction from the AXI master side.
    AxiReadActive :
    begin
      sAxiRDataReg = mAxiRDataReg;
      sAxiRRespReg = mAxiRRespReg;
      if (mAxiRPending & ~sAxiRBlocked)
      begin
        axiState_d = Idle;
        sAxiRPush = 1'b1;
        mAxiRClear = 1'b1;
      end
    end

    // Initiate write transaction on the AXI master side.
    AxiWriteStart :
    begin
      if (~mAxiAWBlocked)
      begin
        axiState_d = AxiWriteData;
        mAxiAWPush = 1'b1;
        sAxiAWClear = 1'b1;
      end
    end

    // Forward write data to the AXI master side.
    AxiWriteData :
    begin
      if (sAxiWPending & ~mAxiWBlocked)
      begin
        axiState_d = AxiWriteActive;
        mAxiWPush = 1'b1;
        sAxiWClear = 1'b1;
      end
    end

    // Complete write transaction from the AXI master side.
    AxiWriteActive :
    begin
      sAxiBRespReg = mAxiBRespReg;
      if (mAxiBPending & ~sAxiBBlocked)
      begin
        axiState_d = Idle;
        sAxiBPush = 1'b1;
        mAxiBClear = 1'b1;
      end
    end

    // Map unknown states to Idle.
    default :
    begin
      axiState_d = Idle;
    end
  endcase
end

// Implement sequential logic for AXI transaction state machine.
always @(posedge clk)
begin
  if (srst)
  begin
    axiState_q <= Idle;
    regReq_q <= 1'b0;
    regWriteEn_q <= 1'b0;
    for (i = 0; i < RegAddrWidth; i = i + 1)
      regAddr_q [i] <= 1'b0;
    regWData_q <= 32'b0;
    regWStrb_q <= 4'b0;
  end
  else
  begin
    axiState_q <= axiState_d;
    regReq_q <= regReq_d;
    regWriteEn_q <= regWriteEn_d;
    regAddr_q <= regAddr_d;
    regWData_q <= regWData_d;
    regWStrb_q <= regWStrb_d;
  end
end

assign regReq = regReq_q;
assign regWriteEn = regWriteEn_q;
assign regAddr = regAddr_q;
assign regWData = regWData_q;
assign regWStrb = regWStrb_q;

endmodule

//
// Provides common implementation of single AXI data input register.
//
// verilator lint_off DECLFILENAME
module sda_kernel_ctrl_reg_sel_axi_inreg_x1
  (axiValid, axiReady, axiDataIn, dataPending, dataClear, dataOut, clk, srst);
// verilator lint_on DECLFILENAME

// Specify the register data width.
parameter DataWidth = 16;

// Specifies the AXI bus input signals.
input                 axiValid;
output                axiReady;
input [DataWidth-1:0] axiDataIn;

// Specifies the data register output signals.
output                 dataPending;
input                  dataClear;
output [DataWidth-1:0] dataOut;

// Specifies the system level signals.
input clk;
input srst;

// Specifies internal registers.
reg                 dataClear_q;
reg                 axiReady_q;
reg [DataWidth-1:0] axiDataIn_q;

integer i;

// Implements a single AXI input register.
always @(posedge clk)
begin
  if (srst | dataClear)
  begin
    dataClear_q <= 1'b1;
    axiReady_q <= 1'b0;
    for (i = 0; i < DataWidth; i = i + 1)
      axiDataIn_q [i] <= 1'b0;
  end
  else if (dataClear_q)
  begin
    dataClear_q <= 1'b0;
    axiReady_q <= 1'b1;
  end
  else if (axiReady_q & axiValid)
  begin
    dataClear_q <= 1'b0;
    axiReady_q <= 1'b0;
    axiDataIn_q <= axiDataIn;
  end
end

assign axiReady = axiReady_q;
assign dataPending = ~(dataClear_q | axiReady_q);
assign dataOut = axiDataIn_q;

endmodule

//
// Provides common implementation of dual AXI data input register.
//
// verilator lint_off DECLFILENAME
module sda_kernel_ctrl_reg_sel_axi_inreg_x2
  (axiValid, axiReady, axiDataIn1, axiDataIn2, dataPending, dataClear,
  dataOut1, dataOut2, clk, srst);
// verilator lint_on DECLFILENAME

// Specify the first register data width.
parameter DataWidth1 = 16;

// Specify the second register data width.
parameter DataWidth2 = 16;

// Specifies the AXI bus input signals.
input                  axiValid;
output                 axiReady;
input [DataWidth1-1:0] axiDataIn1;
input [DataWidth2-1:0] axiDataIn2;

// Specifies the data register output signals.
output                  dataPending;
input                   dataClear;
output [DataWidth1-1:0] dataOut1;
output [DataWidth2-1:0] dataOut2;

// Specifies the system level signals.
input clk;
input srst;

// Specifies the concatenated data vectors.
wire [DataWidth1+DataWidth2-1:0] dataOut;

// Instantiate the single input register module.
sda_kernel_ctrl_reg_sel_axi_inreg_x1 #(DataWidth1+DataWidth2) axiDataReg_u
  (axiValid, axiReady, {axiDataIn2, axiDataIn1}, dataPending, dataClear,
  dataOut, clk, srst);

assign dataOut1 = dataOut [DataWidth1-1:0];
assign dataOut2 = dataOut [DataWidth1+DataWidth2-1:DataWidth1];

endmodule

//
// Provides common implementation of single AXI data output register.
//
// verilator lint_off DECLFILENAME
module sda_kernel_ctrl_reg_sel_axi_outreg_x1
  (dataPush, dataBlocked, dataIn, axiValid, axiReady, axiDataOut, clk, srst);
// verilator lint_on DECLFILENAME

// Specify the register data width.
parameter DataWidth = 16;

// Specifies the data register interface signals.
input                 dataPush;
output                dataBlocked;
input [DataWidth-1:0] dataIn;

// Specifies the AXI bus output signals.
output                 axiValid;
input                  axiReady;
output [DataWidth-1:0] axiDataOut;

// Specifies the system level signals.
input clk;
input srst;

// Specifies internal registers.
reg                 dataReady_q;
reg [DataWidth-1:0] dataReg_q;

integer i;

// Implements a single AXI output register.
always @(posedge clk)
begin
  if (srst)
  begin
    dataReady_q <= 1'b0;
    for (i = 0; i < DataWidth; i = i + 1)
      dataReg_q [i] <= 1'b0;
  end
  else if (dataReady_q & axiReady)
  begin
    dataReady_q <= 1'b0;
  end
  else if (dataPush)
  begin
    dataReady_q <= 1'b1;
    dataReg_q <= dataIn;
  end
end

assign dataBlocked = dataReady_q;
assign axiValid = dataReady_q;
assign axiDataOut = dataReg_q;

endmodule

//
// Provides common implementation of dual AXI data output register.
//
// verilator lint_off DECLFILENAME
module sda_kernel_ctrl_reg_sel_axi_outreg_x2
  (dataPush, dataBlocked, dataIn1, dataIn2, axiValid, axiReady, axiDataOut1,
  axiDataOut2, clk, srst);
// verilator lint_on DECLFILENAME

// Specify the first register data width.
parameter DataWidth1 = 16;

// Specify the second register data width.
parameter DataWidth2 = 16;

// Specifies the data register interface signals.
input                  dataPush;
output                 dataBlocked;
input [DataWidth1-1:0] dataIn1;
input [DataWidth2-1:0] dataIn2;

// Specifies the AXI bus output signals.
output                  axiValid;
input                   axiReady;
output [DataWidth1-1:0] axiDataOut1;
output [DataWidth2-1:0] axiDataOut2;

// Specifies the system level signals.
input clk;
input srst;

// Specifies the concatenated data vectors.
wire [DataWidth1+DataWidth2-1:0] axiDataOut;

// Instantiate the single output register module.
sda_kernel_ctrl_reg_sel_axi_outreg_x1 #(DataWidth1+DataWidth2) axiDataReg_u
  (dataPush, dataBlocked, {dataIn2, dataIn1}, axiValid, axiReady,
  axiDataOut, clk, srst);

assign axiDataOut1 = axiDataOut [DataWidth1-1:0];
assign axiDataOut2 = axiDataOut [DataWidth1+DataWidth2-1:DataWidth1];

endmodule

