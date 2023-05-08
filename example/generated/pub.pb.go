// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.12
// source: pub.proto

package example

import (
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

type HelloWorldEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventID       string `protobuf:"bytes,1,opt,name=eventID,proto3" json:"eventID,omitempty"`
	Name          string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	UnixTimeStamp int64  `protobuf:"varint,3,opt,name=unixTimeStamp,proto3" json:"unixTimeStamp,omitempty"`
}

func (x *HelloWorldEvent) Reset() {
	*x = HelloWorldEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pub_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HelloWorldEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HelloWorldEvent) ProtoMessage() {}

func (x *HelloWorldEvent) ProtoReflect() protoreflect.Message {
	mi := &file_pub_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HelloWorldEvent.ProtoReflect.Descriptor instead.
func (*HelloWorldEvent) Descriptor() ([]byte, []int) {
	return file_pub_proto_rawDescGZIP(), []int{0}
}

func (x *HelloWorldEvent) GetEventID() string {
	if x != nil {
		return x.EventID
	}
	return ""
}

func (x *HelloWorldEvent) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *HelloWorldEvent) GetUnixTimeStamp() int64 {
	if x != nil {
		return x.UnixTimeStamp
	}
	return 0
}

type HogeEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventID       string `protobuf:"bytes,1,opt,name=eventID,proto3" json:"eventID,omitempty"`
	Message       string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	IsHoge        bool   `protobuf:"varint,3,opt,name=isHoge,proto3" json:"isHoge,omitempty"`
	UnixTimeStamp int64  `protobuf:"varint,4,opt,name=unixTimeStamp,proto3" json:"unixTimeStamp,omitempty"`
}

func (x *HogeEvent) Reset() {
	*x = HogeEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pub_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HogeEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HogeEvent) ProtoMessage() {}

func (x *HogeEvent) ProtoReflect() protoreflect.Message {
	mi := &file_pub_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HogeEvent.ProtoReflect.Descriptor instead.
func (*HogeEvent) Descriptor() ([]byte, []int) {
	return file_pub_proto_rawDescGZIP(), []int{1}
}

func (x *HogeEvent) GetEventID() string {
	if x != nil {
		return x.EventID
	}
	return ""
}

func (x *HogeEvent) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *HogeEvent) GetIsHoge() bool {
	if x != nil {
		return x.IsHoge
	}
	return false
}

func (x *HogeEvent) GetUnixTimeStamp() int64 {
	if x != nil {
		return x.UnixTimeStamp
	}
	return 0
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pub_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_pub_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_pub_proto_rawDescGZIP(), []int{2}
}

var File_pub_proto protoreflect.FileDescriptor

var file_pub_proto_rawDesc = []byte{
	0x0a, 0x09, 0x70, 0x75, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x1a, 0x0c, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x65, 0x0a, 0x0f, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x75, 0x6e, 0x69, 0x78, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x75, 0x6e, 0x69, 0x78, 0x54,
	0x69, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x7d, 0x0a, 0x09, 0x48, 0x6f, 0x67, 0x65,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12,
	0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x73, 0x48,
	0x6f, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x73, 0x48, 0x6f, 0x67,
	0x65, 0x12, 0x24, 0x0a, 0x0d, 0x75, 0x6e, 0x69, 0x78, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x74, 0x61,
	0x6d, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x75, 0x6e, 0x69, 0x78, 0x54, 0x69,
	0x6d, 0x65, 0x53, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x32, 0xeb, 0x01, 0x0a, 0x11, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4b, 0x0a, 0x0a, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57,
	0x6f, 0x72, 0x6c, 0x64, 0x12, 0x17, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x48, 0x65,
	0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x1a, 0x0d, 0x2e,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x15, 0xb2, 0xb5,
	0x18, 0x11, 0x0a, 0x0f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x12, 0x42, 0x0a, 0x0b, 0x48, 0x6f, 0x67, 0x65, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x12, 0x11, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x48, 0x6f, 0x67, 0x65,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x1a, 0x0d, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x22, 0x11, 0xb2, 0xb5, 0x18, 0x0d, 0x0a, 0x0b, 0x68, 0x6f, 0x67, 0x65,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x45, 0x0a, 0x0c, 0x48, 0x6f, 0x67, 0x65, 0x73,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x11, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x48, 0x6f, 0x67, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x1a, 0x0d, 0x2e, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x13, 0xb2, 0xb5, 0x18, 0x0f, 0x0a,
	0x0b, 0x68, 0x6f, 0x67, 0x65, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x10, 0x01, 0x42, 0x36,
	0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x61, 0x6b,
	0x61, 0x57, 0x61, 0x6b, 0x69, 0x79, 0x61, 0x6d, 0x61, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2f, 0x65,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pub_proto_rawDescOnce sync.Once
	file_pub_proto_rawDescData = file_pub_proto_rawDesc
)

func file_pub_proto_rawDescGZIP() []byte {
	file_pub_proto_rawDescOnce.Do(func() {
		file_pub_proto_rawDescData = protoimpl.X.CompressGZIP(file_pub_proto_rawDescData)
	})
	return file_pub_proto_rawDescData
}

var file_pub_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pub_proto_goTypes = []interface{}{
	(*HelloWorldEvent)(nil), // 0: events.HelloWorldEvent
	(*HogeEvent)(nil),       // 1: events.HogeEvent
	(*Empty)(nil),           // 2: events.Empty
}
var file_pub_proto_depIdxs = []int32{
	0, // 0: events.HelloWorldService.HelloWorld:input_type -> events.HelloWorldEvent
	1, // 1: events.HelloWorldService.HogeCreated:input_type -> events.HogeEvent
	1, // 2: events.HelloWorldService.HogesCreated:input_type -> events.HogeEvent
	2, // 3: events.HelloWorldService.HelloWorld:output_type -> events.Empty
	2, // 4: events.HelloWorldService.HogeCreated:output_type -> events.Empty
	2, // 5: events.HelloWorldService.HogesCreated:output_type -> events.Empty
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pub_proto_init() }
func file_pub_proto_init() {
	if File_pub_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pub_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HelloWorldEvent); i {
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
		file_pub_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HogeEvent); i {
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
		file_pub_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
			RawDescriptor: file_pub_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pub_proto_goTypes,
		DependencyIndexes: file_pub_proto_depIdxs,
		MessageInfos:      file_pub_proto_msgTypes,
	}.Build()
	File_pub_proto = out.File
	file_pub_proto_rawDesc = nil
	file_pub_proto_goTypes = nil
	file_pub_proto_depIdxs = nil
}
