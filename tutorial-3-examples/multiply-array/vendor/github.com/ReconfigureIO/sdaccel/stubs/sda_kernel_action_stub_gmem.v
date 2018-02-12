//
// (c) 2017 ReconfigureIO
//
// <COPYRIGHT TERMS>
//

//
// Provides a 'stub' implementation of the kernel action logic with single AXI
// shared memory interface.
//

`timescale 1ns/1ps

// Can be redefined on the synthesis command line.
`define AXI_MASTER_ADDR_WIDTH 64

// Can be redefined on the synthesis command line.
`define AXI_MASTER_DATA_WIDTH 64

// Can be redefined on the synthesis command line.
`define AXI_MASTER_ID_WIDTH 1

// Can be redefined on the synthesis command line.
`define AXI_MASTER_USER_WIDTH 1

// The module name is common for different kernel action toplevel entities.
// verilator lint_off DECLFILENAME
module teak__action__top__gmem
  (go_0Ready, go_0Stop, done_0Ready, done_0Stop, s_axi_araddr, s_axi_arcache, s_axi_arprot,
  s_axi_arvalid, s_axi_arready, s_axi_rdata, s_axi_rresp, s_axi_rvalid,
  s_axi_rready, s_axi_awaddr, s_axi_awcache, s_axi_awprot, s_axi_awvalid,
  s_axi_awready, s_axi_wdata, s_axi_wstrb, s_axi_wvalid, s_axi_wready,
  s_axi_bresp, s_axi_bvalid, s_axi_bready, m_axi_gmem_awaddr, m_axi_gmem_awlen,
  m_axi_gmem_awsize, m_axi_gmem_awburst, m_axi_gmem_awlock, m_axi_gmem_awcache,
  m_axi_gmem_awprot, m_axi_gmem_awqos, m_axi_gmem_awregion, m_axi_gmem_awuser,
  m_axi_gmem_awid, m_axi_gmem_awvalid, m_axi_gmem_awready, m_axi_gmem_wdata,
  m_axi_gmem_wstrb, m_axi_gmem_wlast, m_axi_gmem_wuser,
  m_axi_gmem_wvalid, m_axi_gmem_wready, m_axi_gmem_bresp, m_axi_gmem_buser,
  m_axi_gmem_bid, m_axi_gmem_bvalid, m_axi_gmem_bready, m_axi_gmem_araddr,
  m_axi_gmem_arlen, m_axi_gmem_arsize, m_axi_gmem_arburst, m_axi_gmem_arlock,
  m_axi_gmem_arcache, m_axi_gmem_arprot, m_axi_gmem_arqos, m_axi_gmem_arregion,
  m_axi_gmem_aruser, m_axi_gmem_arid, m_axi_gmem_arvalid, m_axi_gmem_arready,
  m_axi_gmem_rdata, m_axi_gmem_rresp, m_axi_gmem_rlast, m_axi_gmem_ruser,
  m_axi_gmem_rid, m_axi_gmem_rvalid, m_axi_gmem_rready, paramaddr_0Ready,
  paramaddr_0Data, paramaddr_0Stop, paramdata_0Ready, paramdata_0Data, paramdata_0Stop,
  clk, reset);
// verilator lint_on DECLFILENAME

// Action control signals.
input  go_0Ready;
output go_0Stop;
output done_0Ready;
input  done_0Stop;

// Parameter data access signals. Provides a SELF channel output for address
// values and a SELF channel input for the corresponding data items read from
// the parameter register file.
// verilator lint_off UNUSED
output        paramaddr_0Ready;
output [31:0] paramaddr_0Data;
input         paramaddr_0Stop;

input         paramdata_0Ready;
input [31:0]  paramdata_0Data;
output        paramdata_0Stop;
// verilator lint_on UNUSED

// AXI interface signals are not used in the stub implementation.
// verilator lint_off UNUSED

// Specifies the AXI slave bus signals.
input [31:0]  s_axi_araddr;
input [3:0]   s_axi_arcache;
input [2:0]   s_axi_arprot;
input         s_axi_arvalid;
output        s_axi_arready;
output [31:0] s_axi_rdata;
output [1:0]  s_axi_rresp;
output        s_axi_rvalid;
input         s_axi_rready;
input [31:0]  s_axi_awaddr;
input [3:0]   s_axi_awcache;
input [2:0]   s_axi_awprot;
input         s_axi_awvalid;
output        s_axi_awready;
input [31:0]  s_axi_wdata;
input [3:0]   s_axi_wstrb;
input         s_axi_wvalid;
output        s_axi_wready;
output [1:0]  s_axi_bresp;
output        s_axi_bvalid;
input         s_axi_bready;

// Specifies the AXI master write address signals.
output [`AXI_MASTER_ADDR_WIDTH-1:0] m_axi_gmem_awaddr;
output [7:0]                        m_axi_gmem_awlen;
output [2:0]                        m_axi_gmem_awsize;
output [1:0]                        m_axi_gmem_awburst;
output                              m_axi_gmem_awlock;
output [3:0]                        m_axi_gmem_awcache;
output [2:0]                        m_axi_gmem_awprot;
output [3:0]                        m_axi_gmem_awqos;
output [3:0]                        m_axi_gmem_awregion;
output [`AXI_MASTER_USER_WIDTH-1:0] m_axi_gmem_awuser;
output [`AXI_MASTER_ID_WIDTH-1:0]   m_axi_gmem_awid;
output                              m_axi_gmem_awvalid;
input                               m_axi_gmem_awready;

// Specifies the AXI master write data signals.
output [`AXI_MASTER_DATA_WIDTH-1:0]   m_axi_gmem_wdata;
output [`AXI_MASTER_DATA_WIDTH/8-1:0] m_axi_gmem_wstrb;
output                                m_axi_gmem_wlast;
output [`AXI_MASTER_USER_WIDTH-1:0]   m_axi_gmem_wuser;
output                                m_axi_gmem_wvalid;
input                                 m_axi_gmem_wready;

// Specifies the AXI master write response signals.
input [1:0]                        m_axi_gmem_bresp;
input [`AXI_MASTER_USER_WIDTH-1:0] m_axi_gmem_buser;
input [`AXI_MASTER_ID_WIDTH-1:0]   m_axi_gmem_bid;
input                              m_axi_gmem_bvalid;
output                             m_axi_gmem_bready;

// Specifies the AXI master read address signals.
output [`AXI_MASTER_ADDR_WIDTH-1:0] m_axi_gmem_araddr;
output [7:0]                        m_axi_gmem_arlen;
output [2:0]                        m_axi_gmem_arsize;
output [1:0]                        m_axi_gmem_arburst;
output                              m_axi_gmem_arlock;
output [3:0]                        m_axi_gmem_arcache;
output [2:0]                        m_axi_gmem_arprot;
output [3:0]                        m_axi_gmem_arqos;
output [3:0]                        m_axi_gmem_arregion;
output [`AXI_MASTER_USER_WIDTH-1:0] m_axi_gmem_aruser;
output [`AXI_MASTER_ID_WIDTH-1:0]   m_axi_gmem_arid;
output                              m_axi_gmem_arvalid;
input                               m_axi_gmem_arready;

// Specifies the AXI master read data signals.
input [`AXI_MASTER_DATA_WIDTH-1:0] m_axi_gmem_rdata;
input [1:0]                        m_axi_gmem_rresp;
input                              m_axi_gmem_rlast;
input [`AXI_MASTER_USER_WIDTH-1:0] m_axi_gmem_ruser;
input [`AXI_MASTER_ID_WIDTH-1:0]   m_axi_gmem_rid;
input                              m_axi_gmem_rvalid;
output                             m_axi_gmem_rready;
// verilator lint_on UNUSED

// System level signals.
input clk;
input reset;

// Action start loopback signals.
reg action_done_q;

// AXI slave loopback signals.
reg s_axi_read_ready_q;
reg s_axi_read_complete_q;
reg s_axi_write_ready_q;
reg s_axi_write_complete_q;

// Implement action start loopback signals.
always @(posedge clk)
begin
  if (reset)
  begin
    action_done_q <= 1'b0;
  end
  else if (action_done_q)
  begin
    action_done_q <= done_0Stop;
  end
  else if (go_0Ready)
  begin
    action_done_q <= 1'b1;
  end
end

assign go_0Stop = action_done_q;
assign done_0Ready = action_done_q;

// Implement AXI read control loopback.
always @(posedge clk)
begin
  if (reset)
  begin
    s_axi_read_ready_q <= 1'b0;
    s_axi_read_complete_q <= 1'b0;
  end
  else if (s_axi_read_complete_q)
  begin
    s_axi_read_complete_q <= ~s_axi_rready;
  end
  else if (s_axi_read_ready_q)
  begin
    s_axi_read_ready_q <= 1'b0;
    s_axi_read_complete_q <= 1'b1;
  end
  else
  begin
    s_axi_read_ready_q <= s_axi_arvalid;
  end
end

assign s_axi_arready = s_axi_read_ready_q;
assign s_axi_rdata = 32'b0;
assign s_axi_rresp = 2'b0;
assign s_axi_rvalid = s_axi_read_complete_q;

// Implement AXI write control loopback.
always @(posedge clk)
begin
  if (reset)
  begin
    s_axi_write_ready_q <= 1'b0;
    s_axi_write_complete_q <= 1'b0;
  end
  else if (s_axi_write_complete_q)
  begin
    s_axi_write_complete_q <= ~s_axi_bready;
  end
  else if (s_axi_write_ready_q)
  begin
    s_axi_write_ready_q <= 1'b0;
    s_axi_write_complete_q <= 1'b1;
  end
  else
  begin
    s_axi_write_ready_q <= s_axi_awvalid & s_axi_wvalid;
  end
end

assign s_axi_awready = s_axi_write_ready_q;
assign s_axi_wready = s_axi_write_ready_q;
assign s_axi_bresp = 2'b0;
assign s_axi_bvalid = s_axi_write_complete_q;

// Tie off unused parameter access signals.
assign paramaddr_0Ready = 1'b0;
assign paramaddr_0Data = 32'b0;
assign paramdata_0Stop = 1'b0;

// Tie of unused AXI memory access signals.
assign m_axi_gmem_awaddr = `AXI_MASTER_ADDR_WIDTH'b0;
assign m_axi_gmem_awlen = 8'b0;
assign m_axi_gmem_awsize = 3'b0;
assign m_axi_gmem_awburst = 2'b0;
assign m_axi_gmem_awlock = 1'b0;
assign m_axi_gmem_awcache = 4'b0;
assign m_axi_gmem_awprot = 3'b0;
assign m_axi_gmem_awqos = 4'b0;
assign m_axi_gmem_awregion = 4'b0;
assign m_axi_gmem_awuser = `AXI_MASTER_USER_WIDTH'b0;
assign m_axi_gmem_awid = `AXI_MASTER_ID_WIDTH'b0;
assign m_axi_gmem_awvalid = 1'b0;
assign m_axi_gmem_wdata = `AXI_MASTER_DATA_WIDTH'b0;
assign m_axi_gmem_wstrb = `AXI_MASTER_DATA_WIDTH/8'b0;
assign m_axi_gmem_wlast = 1'b0;
assign m_axi_gmem_wuser = `AXI_MASTER_USER_WIDTH'b0;
assign m_axi_gmem_wvalid = 1'b0;
assign m_axi_gmem_bready = 1'b0;
assign m_axi_gmem_araddr = `AXI_MASTER_ADDR_WIDTH'b0;
assign m_axi_gmem_arlen = 8'b0;
assign m_axi_gmem_arsize = 3'b0;
assign m_axi_gmem_arburst = 2'b0;
assign m_axi_gmem_arlock = 1'b0;
assign m_axi_gmem_arcache = 4'b0;
assign m_axi_gmem_arprot = 3'b0;
assign m_axi_gmem_arqos = 4'b0;
assign m_axi_gmem_arregion = 4'b0;
assign m_axi_gmem_aruser = `AXI_MASTER_USER_WIDTH'b0;
assign m_axi_gmem_arid = `AXI_MASTER_ID_WIDTH'b0;
assign m_axi_gmem_arvalid = 1'b0;
assign m_axi_gmem_rready = 1'b0;

endmodule
