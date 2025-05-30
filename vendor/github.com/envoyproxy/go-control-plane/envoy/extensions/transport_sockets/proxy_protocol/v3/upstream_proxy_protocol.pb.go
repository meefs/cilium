// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v5.29.3
// source: envoy/extensions/transport_sockets/proxy_protocol/v3/upstream_proxy_protocol.proto

package proxy_protocolv3

import (
	_ "github.com/cncf/xds/go/udpa/annotations"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Configuration for PROXY protocol socket
type ProxyProtocolUpstreamTransport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The PROXY protocol settings
	Config *v3.ProxyProtocolConfig `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	// The underlying transport socket being wrapped.
	TransportSocket *v3.TransportSocket `protobuf:"bytes,2,opt,name=transport_socket,json=transportSocket,proto3" json:"transport_socket,omitempty"`
	// If this is set to true, the null addresses are allowed in the PROXY protocol header.
	// The proxy protocol header encodes the null addresses to AF_UNSPEC.
	// [#not-implemented-hide:]
	AllowUnspecifiedAddress bool `protobuf:"varint,3,opt,name=allow_unspecified_address,json=allowUnspecifiedAddress,proto3" json:"allow_unspecified_address,omitempty"`
	// If true, all the TLVs are encoded in the connection pool key.
	// [#not-implemented-hide:]
	TlvAsPoolKey bool `protobuf:"varint,4,opt,name=tlv_as_pool_key,json=tlvAsPoolKey,proto3" json:"tlv_as_pool_key,omitempty"`
}

func (x *ProxyProtocolUpstreamTransport) Reset() {
	*x = ProxyProtocolUpstreamTransport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProxyProtocolUpstreamTransport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProxyProtocolUpstreamTransport) ProtoMessage() {}

func (x *ProxyProtocolUpstreamTransport) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProxyProtocolUpstreamTransport.ProtoReflect.Descriptor instead.
func (*ProxyProtocolUpstreamTransport) Descriptor() ([]byte, []int) {
	return file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *ProxyProtocolUpstreamTransport) GetConfig() *v3.ProxyProtocolConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ProxyProtocolUpstreamTransport) GetTransportSocket() *v3.TransportSocket {
	if x != nil {
		return x.TransportSocket
	}
	return nil
}

func (x *ProxyProtocolUpstreamTransport) GetAllowUnspecifiedAddress() bool {
	if x != nil {
		return x.AllowUnspecifiedAddress
	}
	return false
}

func (x *ProxyProtocolUpstreamTransport) GetTlvAsPoolKey() bool {
	if x != nil {
		return x.TlvAsPoolKey
	}
	return false
}

var File_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto protoreflect.FileDescriptor

var file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDesc = []byte{
	0x0a, 0x52, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x2f, 0x76, 0x33, 0x2f, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x34, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65,
	0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74,
	0x5f, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x76, 0x33, 0x1a, 0x1f, 0x65, 0x6e, 0x76, 0x6f,
	0x79, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x33,
	0x2f, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x29, 0x65, 0x6e, 0x76,
	0x6f, 0x79, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76,
	0x33, 0x2f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x75, 0x64, 0x70, 0x61, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa2,
	0x02, 0x0a, 0x1e, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x12, 0x41, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x29, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x33, 0x2e, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x5a, 0x0a, 0x10, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x5f, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x76, 0x33, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x53,
	0x6f, 0x63, 0x6b, 0x65, 0x74, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02, 0x10, 0x01, 0x52,
	0x0f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74,
	0x12, 0x3a, 0x0a, 0x19, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x75, 0x6e, 0x73, 0x70, 0x65, 0x63,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x17, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x63,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x25, 0x0a, 0x0f,
	0x74, 0x6c, 0x76, 0x5f, 0x61, 0x73, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x5f, 0x6b, 0x65, 0x79, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x74, 0x6c, 0x76, 0x41, 0x73, 0x50, 0x6f, 0x6f, 0x6c,
	0x4b, 0x65, 0x79, 0x42, 0xd8, 0x01, 0xba, 0x80, 0xc8, 0xd1, 0x06, 0x02, 0x10, 0x02, 0x0a, 0x42,
	0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x65, 0x6e,
	0x76, 0x6f, 0x79, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e,
	0x76, 0x33, 0x42, 0x1a, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x50, 0x72, 0x6f, 0x78,
	0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x6c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76,
	0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x2d, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x65,
	0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x5f, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x76, 0x33, 0x3b, 0x70, 0x72,
	0x6f, 0x78, 0x79, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x76, 0x33, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescOnce sync.Once
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescData = file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDesc
)

func file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescGZIP() []byte {
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescOnce.Do(func() {
		file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescData)
	})
	return file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDescData
}

var file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_goTypes = []interface{}{
	(*ProxyProtocolUpstreamTransport)(nil), // 0: envoy.extensions.transport_sockets.proxy_protocol.v3.ProxyProtocolUpstreamTransport
	(*v3.ProxyProtocolConfig)(nil),         // 1: envoy.config.core.v3.ProxyProtocolConfig
	(*v3.TransportSocket)(nil),             // 2: envoy.config.core.v3.TransportSocket
}
var file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_depIdxs = []int32{
	1, // 0: envoy.extensions.transport_sockets.proxy_protocol.v3.ProxyProtocolUpstreamTransport.config:type_name -> envoy.config.core.v3.ProxyProtocolConfig
	2, // 1: envoy.extensions.transport_sockets.proxy_protocol.v3.ProxyProtocolUpstreamTransport.transport_socket:type_name -> envoy.config.core.v3.TransportSocket
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() {
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_init()
}
func file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_init() {
	if File_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProxyProtocolUpstreamTransport); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_goTypes,
		DependencyIndexes: file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_depIdxs,
		MessageInfos:      file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_msgTypes,
	}.Build()
	File_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto = out.File
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_rawDesc = nil
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_goTypes = nil
	file_envoy_extensions_transport_sockets_proxy_protocol_v3_upstream_proxy_protocol_proto_depIdxs = nil
}
