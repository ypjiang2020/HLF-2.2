// Code generated by protoc-gen-go. DO NOT EDIT.
// source: peer/snapshot.proto

package peer

import (
	context "context"
	fmt "fmt"
	common "github.com/Yunpeng-J/fabric-protos-go/common"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// SnapshotRequest contains information for a generate/cancel snapshot request
type SnapshotRequest struct {
	// The signature header that contains creator identity and nonce
	SignatureHeader *common.SignatureHeader `protobuf:"bytes,1,opt,name=signature_header,json=signatureHeader,proto3" json:"signature_header,omitempty"`
	// The channel ID
	ChannelId string `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	// The block number to generate a snapshot
	BlockNumber          uint64   `protobuf:"varint,3,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SnapshotRequest) Reset()         { *m = SnapshotRequest{} }
func (m *SnapshotRequest) String() string { return proto.CompactTextString(m) }
func (*SnapshotRequest) ProtoMessage()    {}
func (*SnapshotRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d05a247df97d1516, []int{0}
}

func (m *SnapshotRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnapshotRequest.Unmarshal(m, b)
}
func (m *SnapshotRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnapshotRequest.Marshal(b, m, deterministic)
}
func (m *SnapshotRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotRequest.Merge(m, src)
}
func (m *SnapshotRequest) XXX_Size() int {
	return xxx_messageInfo_SnapshotRequest.Size(m)
}
func (m *SnapshotRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotRequest proto.InternalMessageInfo

func (m *SnapshotRequest) GetSignatureHeader() *common.SignatureHeader {
	if m != nil {
		return m.SignatureHeader
	}
	return nil
}

func (m *SnapshotRequest) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func (m *SnapshotRequest) GetBlockNumber() uint64 {
	if m != nil {
		return m.BlockNumber
	}
	return 0
}

// SnapshotQuery contains information for a query snapshot request
type SnapshotQuery struct {
	// The signature header that contains creator identity and nonce
	SignatureHeader *common.SignatureHeader `protobuf:"bytes,1,opt,name=signature_header,json=signatureHeader,proto3" json:"signature_header,omitempty"`
	// The channel ID
	ChannelId            string   `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SnapshotQuery) Reset()         { *m = SnapshotQuery{} }
func (m *SnapshotQuery) String() string { return proto.CompactTextString(m) }
func (*SnapshotQuery) ProtoMessage()    {}
func (*SnapshotQuery) Descriptor() ([]byte, []int) {
	return fileDescriptor_d05a247df97d1516, []int{1}
}

func (m *SnapshotQuery) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnapshotQuery.Unmarshal(m, b)
}
func (m *SnapshotQuery) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnapshotQuery.Marshal(b, m, deterministic)
}
func (m *SnapshotQuery) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotQuery.Merge(m, src)
}
func (m *SnapshotQuery) XXX_Size() int {
	return xxx_messageInfo_SnapshotQuery.Size(m)
}
func (m *SnapshotQuery) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotQuery.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotQuery proto.InternalMessageInfo

func (m *SnapshotQuery) GetSignatureHeader() *common.SignatureHeader {
	if m != nil {
		return m.SignatureHeader
	}
	return nil
}

func (m *SnapshotQuery) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

// SignedSnapshotRequest contains marshalled request bytes and signature
type SignedSnapshotRequest struct {
	// The bytes of SnapshotRequest or SnapshotQuery
	Request []byte `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	// Signaure over request bytes; this signature is to be verified against the client identity
	Signature            []byte   `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SignedSnapshotRequest) Reset()         { *m = SignedSnapshotRequest{} }
func (m *SignedSnapshotRequest) String() string { return proto.CompactTextString(m) }
func (*SignedSnapshotRequest) ProtoMessage()    {}
func (*SignedSnapshotRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d05a247df97d1516, []int{2}
}

func (m *SignedSnapshotRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignedSnapshotRequest.Unmarshal(m, b)
}
func (m *SignedSnapshotRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignedSnapshotRequest.Marshal(b, m, deterministic)
}
func (m *SignedSnapshotRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignedSnapshotRequest.Merge(m, src)
}
func (m *SignedSnapshotRequest) XXX_Size() int {
	return xxx_messageInfo_SignedSnapshotRequest.Size(m)
}
func (m *SignedSnapshotRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SignedSnapshotRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SignedSnapshotRequest proto.InternalMessageInfo

func (m *SignedSnapshotRequest) GetRequest() []byte {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *SignedSnapshotRequest) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

// QueryPendingSnapshotsResponse specifies the response payload of a query pending snapshots request
type QueryPendingSnapshotsResponse struct {
	BlockNumbers         []uint64 `protobuf:"varint,1,rep,packed,name=block_numbers,json=blockNumbers,proto3" json:"block_numbers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryPendingSnapshotsResponse) Reset()         { *m = QueryPendingSnapshotsResponse{} }
func (m *QueryPendingSnapshotsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryPendingSnapshotsResponse) ProtoMessage()    {}
func (*QueryPendingSnapshotsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d05a247df97d1516, []int{3}
}

func (m *QueryPendingSnapshotsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryPendingSnapshotsResponse.Unmarshal(m, b)
}
func (m *QueryPendingSnapshotsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryPendingSnapshotsResponse.Marshal(b, m, deterministic)
}
func (m *QueryPendingSnapshotsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPendingSnapshotsResponse.Merge(m, src)
}
func (m *QueryPendingSnapshotsResponse) XXX_Size() int {
	return xxx_messageInfo_QueryPendingSnapshotsResponse.Size(m)
}
func (m *QueryPendingSnapshotsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPendingSnapshotsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPendingSnapshotsResponse proto.InternalMessageInfo

func (m *QueryPendingSnapshotsResponse) GetBlockNumbers() []uint64 {
	if m != nil {
		return m.BlockNumbers
	}
	return nil
}

func init() {
	proto.RegisterType((*SnapshotRequest)(nil), "protos.SnapshotRequest")
	proto.RegisterType((*SnapshotQuery)(nil), "protos.SnapshotQuery")
	proto.RegisterType((*SignedSnapshotRequest)(nil), "protos.SignedSnapshotRequest")
	proto.RegisterType((*QueryPendingSnapshotsResponse)(nil), "protos.QueryPendingSnapshotsResponse")
}

func init() { proto.RegisterFile("peer/snapshot.proto", fileDescriptor_d05a247df97d1516) }

var fileDescriptor_d05a247df97d1516 = []byte{
	// 398 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x53, 0xcb, 0x8e, 0xd3, 0x30,
	0x14, 0x9d, 0x30, 0xa3, 0x61, 0x7a, 0x27, 0x55, 0x91, 0x2b, 0x20, 0x2a, 0x54, 0x0a, 0x41, 0x48,
	0x59, 0x50, 0x47, 0x2a, 0x5f, 0x40, 0x0b, 0x02, 0x36, 0x3c, 0xdc, 0x05, 0x12, 0x9b, 0x2a, 0x8f,
	0x5b, 0x27, 0x22, 0xb1, 0x83, 0x9d, 0x2c, 0xfa, 0x25, 0x7c, 0x24, 0x3f, 0x81, 0x62, 0xd7, 0x10,
	0x10, 0x02, 0x89, 0xc5, 0xac, 0x1c, 0x9f, 0x7b, 0x7c, 0x7c, 0x72, 0x7c, 0x2f, 0xcc, 0x5b, 0x44,
	0x95, 0x68, 0x91, 0xb6, 0xba, 0x94, 0x1d, 0x6d, 0x95, 0xec, 0x24, 0xb9, 0x34, 0x8b, 0x5e, 0x3c,
	0xe0, 0x52, 0xf2, 0x1a, 0x13, 0xb3, 0xcd, 0xfa, 0x43, 0x82, 0x4d, 0xdb, 0x1d, 0x2d, 0x69, 0x31,
	0xcf, 0x65, 0xd3, 0x48, 0x91, 0xd8, 0xc5, 0x82, 0xd1, 0x57, 0x0f, 0x66, 0xbb, 0x93, 0x18, 0xc3,
	0x2f, 0x3d, 0xea, 0x8e, 0x6c, 0xe0, 0x8e, 0xae, 0xb8, 0x48, 0xbb, 0x5e, 0xe1, 0xbe, 0xc4, 0xb4,
	0x40, 0x15, 0x78, 0xa1, 0x17, 0x5f, 0xaf, 0xef, 0xd3, 0xd3, 0xe1, 0x9d, 0xab, 0xbf, 0x36, 0x65,
	0x36, 0xd3, 0xbf, 0x02, 0x64, 0x09, 0x90, 0x97, 0xa9, 0x10, 0x58, 0xef, 0xab, 0x22, 0xb8, 0x15,
	0x7a, 0xf1, 0x84, 0x4d, 0x4e, 0xc8, 0x9b, 0x82, 0x3c, 0x02, 0x3f, 0xab, 0x65, 0xfe, 0x79, 0x2f,
	0xfa, 0x26, 0x43, 0x15, 0x9c, 0x87, 0x5e, 0x7c, 0xc1, 0xae, 0x0d, 0xf6, 0xd6, 0x40, 0x91, 0x82,
	0xa9, 0x33, 0xf6, 0xa1, 0x47, 0x75, 0xbc, 0x01, 0x5b, 0xd1, 0x3b, 0xb8, 0x3b, 0x48, 0x60, 0xf1,
	0x7b, 0x24, 0x01, 0xdc, 0x56, 0xf6, 0xd3, 0x5c, 0xe9, 0x33, 0xb7, 0x25, 0x0f, 0x61, 0xf2, 0xe3,
	0x12, 0x23, 0xe8, 0xb3, 0x9f, 0x40, 0xf4, 0x02, 0x96, 0xc6, 0xfc, 0x7b, 0x14, 0x45, 0x25, 0xb8,
	0x93, 0xd5, 0x0c, 0x75, 0x2b, 0x85, 0x46, 0xf2, 0x18, 0xa6, 0xe3, 0x20, 0x74, 0xe0, 0x85, 0xe7,
	0xf1, 0x05, 0xf3, 0x47, 0x49, 0xe8, 0xf5, 0x37, 0x0f, 0xae, 0xdc, 0x51, 0xb2, 0x85, 0xab, 0x57,
	0x28, 0x50, 0xa5, 0x1d, 0x92, 0xa5, 0x7d, 0x45, 0x4d, 0xff, 0xe8, 0x7a, 0x71, 0x8f, 0xda, 0x7e,
	0xa0, 0xae, 0x1f, 0xe8, 0xcb, 0xa1, 0x1f, 0xa2, 0x33, 0xf2, 0x1c, 0x2e, 0xb7, 0xa9, 0xc8, 0xb1,
	0xfe, 0x7f, 0x89, 0x8f, 0x30, 0x1d, 0xff, 0x9a, 0xfe, 0x97, 0xd2, 0x13, 0x57, 0xfe, 0x6b, 0x20,
	0xd1, 0xd9, 0x86, 0x41, 0x24, 0x15, 0xa7, 0xe5, 0xb1, 0x45, 0x55, 0x63, 0xc1, 0x51, 0xd1, 0x43,
	0x9a, 0xa9, 0x2a, 0x77, 0x02, 0xc3, 0x04, 0x7c, 0x7a, 0xca, 0xab, 0xae, 0xec, 0xb3, 0xe1, 0xe5,
	0x93, 0x11, 0x35, 0xb1, 0xd4, 0x95, 0xa5, 0xae, 0xb8, 0x4c, 0x06, 0x76, 0x66, 0x07, 0xe4, 0xd9,
	0xf7, 0x00, 0x00, 0x00, 0xff, 0xff, 0x2f, 0x39, 0x38, 0xb5, 0x3e, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SnapshotClient is the client API for Snapshot service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SnapshotClient interface {
	// Generate a snapshot reqeust. SignedSnapshotRequest contains marshalled bytes for SnaphostRequest
	Generate(ctx context.Context, in *SignedSnapshotRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Cancel a snapshot reqeust. SignedSnapshotRequest contains marshalled bytes for SnaphostRequest
	Cancel(ctx context.Context, in *SignedSnapshotRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Query pending snapshots query. SignedSnapshotRequest contains marshalled bytes for SnaphostQuery
	QueryPendings(ctx context.Context, in *SignedSnapshotRequest, opts ...grpc.CallOption) (*QueryPendingSnapshotsResponse, error)
}

type snapshotClient struct {
	cc *grpc.ClientConn
}

func NewSnapshotClient(cc *grpc.ClientConn) SnapshotClient {
	return &snapshotClient{cc}
}

func (c *snapshotClient) Generate(ctx context.Context, in *SignedSnapshotRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/protos.Snapshot/Generate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *snapshotClient) Cancel(ctx context.Context, in *SignedSnapshotRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/protos.Snapshot/Cancel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *snapshotClient) QueryPendings(ctx context.Context, in *SignedSnapshotRequest, opts ...grpc.CallOption) (*QueryPendingSnapshotsResponse, error) {
	out := new(QueryPendingSnapshotsResponse)
	err := c.cc.Invoke(ctx, "/protos.Snapshot/QueryPendings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SnapshotServer is the server API for Snapshot service.
type SnapshotServer interface {
	// Generate a snapshot reqeust. SignedSnapshotRequest contains marshalled bytes for SnaphostRequest
	Generate(context.Context, *SignedSnapshotRequest) (*empty.Empty, error)
	// Cancel a snapshot reqeust. SignedSnapshotRequest contains marshalled bytes for SnaphostRequest
	Cancel(context.Context, *SignedSnapshotRequest) (*empty.Empty, error)
	// Query pending snapshots query. SignedSnapshotRequest contains marshalled bytes for SnaphostQuery
	QueryPendings(context.Context, *SignedSnapshotRequest) (*QueryPendingSnapshotsResponse, error)
}

// UnimplementedSnapshotServer can be embedded to have forward compatible implementations.
type UnimplementedSnapshotServer struct {
}

func (*UnimplementedSnapshotServer) Generate(ctx context.Context, req *SignedSnapshotRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Generate not implemented")
}
func (*UnimplementedSnapshotServer) Cancel(ctx context.Context, req *SignedSnapshotRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Cancel not implemented")
}
func (*UnimplementedSnapshotServer) QueryPendings(ctx context.Context, req *SignedSnapshotRequest) (*QueryPendingSnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryPendings not implemented")
}

func RegisterSnapshotServer(s *grpc.Server, srv SnapshotServer) {
	s.RegisterService(&_Snapshot_serviceDesc, srv)
}

func _Snapshot_Generate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignedSnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnapshotServer).Generate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Snapshot/Generate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnapshotServer).Generate(ctx, req.(*SignedSnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Snapshot_Cancel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignedSnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnapshotServer).Cancel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Snapshot/Cancel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnapshotServer).Cancel(ctx, req.(*SignedSnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Snapshot_QueryPendings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignedSnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnapshotServer).QueryPendings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Snapshot/QueryPendings",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnapshotServer).QueryPendings(ctx, req.(*SignedSnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Snapshot_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Snapshot",
	HandlerType: (*SnapshotServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Generate",
			Handler:    _Snapshot_Generate_Handler,
		},
		{
			MethodName: "Cancel",
			Handler:    _Snapshot_Cancel_Handler,
		},
		{
			MethodName: "QueryPendings",
			Handler:    _Snapshot_QueryPendings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "peer/snapshot.proto",
}
