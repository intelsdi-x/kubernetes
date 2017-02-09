/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by protoc-gen-gogo.
// source: api.proto
// DO NOT EDIT!

/*
	Package lifecycle is a generated protocol buffer package.

	It is generated from these files:
		api.proto

	It has these top-level messages:
		RegisterRequest
		RegisterReply
		Event
		EventReply
		CgroupInfo
*/
package lifecycle

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.GoGoProtoPackageIsVersion1

type CgroupInfo_Kind int32

const (
	CgroupInfo_QOS       CgroupInfo_Kind = 0
	CgroupInfo_POD       CgroupInfo_Kind = 1
	CgroupInfo_CONTAINER CgroupInfo_Kind = 2
)

var CgroupInfo_Kind_name = map[int32]string{
	0: "QOS",
	1: "POD",
	2: "CONTAINER",
}
var CgroupInfo_Kind_value = map[string]int32{
	"QOS":       0,
	"POD":       1,
	"CONTAINER": 2,
}

func (x CgroupInfo_Kind) String() string {
	return proto.EnumName(CgroupInfo_Kind_name, int32(x))
}
func (CgroupInfo_Kind) EnumDescriptor() ([]byte, []int) { return fileDescriptorApi, []int{4, 0} }

type RegisterRequest struct {
	// For example: localhost:10321
	SocketAddress string `protobuf:"bytes,1,opt,name=socketAddress,proto3" json:"socketAddress,omitempty"`
	Name          string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *RegisterRequest) Reset()                    { *m = RegisterRequest{} }
func (*RegisterRequest) ProtoMessage()               {}
func (*RegisterRequest) Descriptor() ([]byte, []int) { return fileDescriptorApi, []int{0} }

type RegisterReply struct {
	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (m *RegisterReply) Reset()                    { *m = RegisterReply{} }
func (*RegisterReply) ProtoMessage()               {}
func (*RegisterReply) Descriptor() ([]byte, []int) { return fileDescriptorApi, []int{1} }

type Event struct {
	// TODO(CD): Represent event type as an enum.
	Name       string      `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	CgroupInfo *CgroupInfo `protobuf:"bytes,2,opt,name=cgroup_info,json=cgroupInfo" json:"cgroup_info,omitempty"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptorApi, []int{2} }

func (m *Event) GetCgroupInfo() *CgroupInfo {
	if m != nil {
		return m.CgroupInfo
	}
	return nil
}

type EventReply struct {
	Error      string      `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	CgroupInfo *CgroupInfo `protobuf:"bytes,2,opt,name=cgroup_info,json=cgroupInfo" json:"cgroup_info,omitempty"`
}

func (m *EventReply) Reset()                    { *m = EventReply{} }
func (*EventReply) ProtoMessage()               {}
func (*EventReply) Descriptor() ([]byte, []int) { return fileDescriptorApi, []int{3} }

func (m *EventReply) GetCgroupInfo() *CgroupInfo {
	if m != nil {
		return m.CgroupInfo
	}
	return nil
}

type CgroupInfo struct {
	Kind CgroupInfo_Kind `protobuf:"varint,1,opt,name=kind,proto3,enum=lifecycle.CgroupInfo_Kind" json:"kind,omitempty"`
	Path string          `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
}

func (m *CgroupInfo) Reset()                    { *m = CgroupInfo{} }
func (*CgroupInfo) ProtoMessage()               {}
func (*CgroupInfo) Descriptor() ([]byte, []int) { return fileDescriptorApi, []int{4} }

func init() {
	proto.RegisterType((*RegisterRequest)(nil), "lifecycle.RegisterRequest")
	proto.RegisterType((*RegisterReply)(nil), "lifecycle.RegisterReply")
	proto.RegisterType((*Event)(nil), "lifecycle.Event")
	proto.RegisterType((*EventReply)(nil), "lifecycle.EventReply")
	proto.RegisterType((*CgroupInfo)(nil), "lifecycle.CgroupInfo")
	proto.RegisterEnum("lifecycle.CgroupInfo_Kind", CgroupInfo_Kind_name, CgroupInfo_Kind_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for KubeletEventDispatcher service

type KubeletEventDispatcherClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error)
}

type kubeletEventDispatcherClient struct {
	cc *grpc.ClientConn
}

func NewKubeletEventDispatcherClient(cc *grpc.ClientConn) KubeletEventDispatcherClient {
	return &kubeletEventDispatcherClient{cc}
}

func (c *kubeletEventDispatcherClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error) {
	out := new(RegisterReply)
	err := grpc.Invoke(ctx, "/lifecycle.KubeletEventDispatcher/Register", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for KubeletEventDispatcher service

type KubeletEventDispatcherServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterReply, error)
}

func RegisterKubeletEventDispatcherServer(s *grpc.Server, srv KubeletEventDispatcherServer) {
	s.RegisterService(&_KubeletEventDispatcher_serviceDesc, srv)
}

func _KubeletEventDispatcher_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KubeletEventDispatcherServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lifecycle.KubeletEventDispatcher/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KubeletEventDispatcherServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KubeletEventDispatcher_serviceDesc = grpc.ServiceDesc{
	ServiceName: "lifecycle.KubeletEventDispatcher",
	HandlerType: (*KubeletEventDispatcherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _KubeletEventDispatcher_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptorApi,
}

// Client API for LifecycleEventHandler service

type LifecycleEventHandlerClient interface {
	Notify(ctx context.Context, in *Event, opts ...grpc.CallOption) (*EventReply, error)
}

type lifecycleEventHandlerClient struct {
	cc *grpc.ClientConn
}

func NewLifecycleEventHandlerClient(cc *grpc.ClientConn) LifecycleEventHandlerClient {
	return &lifecycleEventHandlerClient{cc}
}

func (c *lifecycleEventHandlerClient) Notify(ctx context.Context, in *Event, opts ...grpc.CallOption) (*EventReply, error) {
	out := new(EventReply)
	err := grpc.Invoke(ctx, "/lifecycle.LifecycleEventHandler/Notify", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LifecycleEventHandler service

type LifecycleEventHandlerServer interface {
	Notify(context.Context, *Event) (*EventReply, error)
}

func RegisterLifecycleEventHandlerServer(s *grpc.Server, srv LifecycleEventHandlerServer) {
	s.RegisterService(&_LifecycleEventHandler_serviceDesc, srv)
}

func _LifecycleEventHandler_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Event)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LifecycleEventHandlerServer).Notify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lifecycle.LifecycleEventHandler/Notify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LifecycleEventHandlerServer).Notify(ctx, req.(*Event))
	}
	return interceptor(ctx, in, info, handler)
}

var _LifecycleEventHandler_serviceDesc = grpc.ServiceDesc{
	ServiceName: "lifecycle.LifecycleEventHandler",
	HandlerType: (*LifecycleEventHandlerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Notify",
			Handler:    _LifecycleEventHandler_Notify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptorApi,
}

func (m *RegisterRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *RegisterRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.SocketAddress) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintApi(data, i, uint64(len(m.SocketAddress)))
		i += copy(data[i:], m.SocketAddress)
	}
	if len(m.Name) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintApi(data, i, uint64(len(m.Name)))
		i += copy(data[i:], m.Name)
	}
	return i, nil
}

func (m *RegisterReply) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *RegisterReply) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Error) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintApi(data, i, uint64(len(m.Error)))
		i += copy(data[i:], m.Error)
	}
	return i, nil
}

func (m *Event) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *Event) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintApi(data, i, uint64(len(m.Name)))
		i += copy(data[i:], m.Name)
	}
	if m.CgroupInfo != nil {
		data[i] = 0x12
		i++
		i = encodeVarintApi(data, i, uint64(m.CgroupInfo.Size()))
		n1, err := m.CgroupInfo.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	return i, nil
}

func (m *EventReply) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *EventReply) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Error) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintApi(data, i, uint64(len(m.Error)))
		i += copy(data[i:], m.Error)
	}
	if m.CgroupInfo != nil {
		data[i] = 0x12
		i++
		i = encodeVarintApi(data, i, uint64(m.CgroupInfo.Size()))
		n2, err := m.CgroupInfo.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *CgroupInfo) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *CgroupInfo) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Kind != 0 {
		data[i] = 0x8
		i++
		i = encodeVarintApi(data, i, uint64(m.Kind))
	}
	if len(m.Path) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintApi(data, i, uint64(len(m.Path)))
		i += copy(data[i:], m.Path)
	}
	return i, nil
}

func encodeFixed64Api(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Api(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintApi(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *RegisterRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.SocketAddress)
	if l > 0 {
		n += 1 + l + sovApi(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovApi(uint64(l))
	}
	return n
}

func (m *RegisterReply) Size() (n int) {
	var l int
	_ = l
	l = len(m.Error)
	if l > 0 {
		n += 1 + l + sovApi(uint64(l))
	}
	return n
}

func (m *Event) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovApi(uint64(l))
	}
	if m.CgroupInfo != nil {
		l = m.CgroupInfo.Size()
		n += 1 + l + sovApi(uint64(l))
	}
	return n
}

func (m *EventReply) Size() (n int) {
	var l int
	_ = l
	l = len(m.Error)
	if l > 0 {
		n += 1 + l + sovApi(uint64(l))
	}
	if m.CgroupInfo != nil {
		l = m.CgroupInfo.Size()
		n += 1 + l + sovApi(uint64(l))
	}
	return n
}

func (m *CgroupInfo) Size() (n int) {
	var l int
	_ = l
	if m.Kind != 0 {
		n += 1 + sovApi(uint64(m.Kind))
	}
	l = len(m.Path)
	if l > 0 {
		n += 1 + l + sovApi(uint64(l))
	}
	return n
}

func sovApi(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozApi(x uint64) (n int) {
	return sovApi(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *RegisterRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&RegisterRequest{`,
		`SocketAddress:` + fmt.Sprintf("%v", this.SocketAddress) + `,`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`}`,
	}, "")
	return s
}
func (this *RegisterReply) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&RegisterReply{`,
		`Error:` + fmt.Sprintf("%v", this.Error) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Event) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Event{`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`CgroupInfo:` + strings.Replace(fmt.Sprintf("%v", this.CgroupInfo), "CgroupInfo", "CgroupInfo", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *EventReply) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&EventReply{`,
		`Error:` + fmt.Sprintf("%v", this.Error) + `,`,
		`CgroupInfo:` + strings.Replace(fmt.Sprintf("%v", this.CgroupInfo), "CgroupInfo", "CgroupInfo", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *CgroupInfo) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&CgroupInfo{`,
		`Kind:` + fmt.Sprintf("%v", this.Kind) + `,`,
		`Path:` + fmt.Sprintf("%v", this.Path) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringApi(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *RegisterRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RegisterRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RegisterRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SocketAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SocketAddress = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipApi(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RegisterReply) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RegisterReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RegisterReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Error", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Error = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipApi(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Event) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Event: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Event: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CgroupInfo", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.CgroupInfo == nil {
				m.CgroupInfo = &CgroupInfo{}
			}
			if err := m.CgroupInfo.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipApi(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventReply) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Error", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Error = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CgroupInfo", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.CgroupInfo == nil {
				m.CgroupInfo = &CgroupInfo{}
			}
			if err := m.CgroupInfo.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipApi(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CgroupInfo) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CgroupInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CgroupInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Kind", wireType)
			}
			m.Kind = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Kind |= (CgroupInfo_Kind(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Path", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Path = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipApi(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipApi(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowApi
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowApi
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowApi
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthApi
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowApi
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipApi(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthApi = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowApi   = fmt.Errorf("proto: integer overflow")
)

var fileDescriptorApi = []byte{
	// 397 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x92, 0x4d, 0xeb, 0xd3, 0x40,
	0x10, 0xc6, 0xb3, 0x7f, 0xdb, 0x6a, 0xa6, 0x54, 0xc3, 0x62, 0xa5, 0x04, 0x09, 0x12, 0x14, 0xbd,
	0x98, 0x42, 0x0a, 0xde, 0xfb, 0x06, 0x96, 0x96, 0x56, 0x53, 0x4f, 0x22, 0x48, 0x5e, 0x36, 0xe9,
	0xd2, 0x34, 0x1b, 0x37, 0x1b, 0x21, 0x37, 0x3f, 0x82, 0x1f, 0xab, 0x47, 0x8f, 0x1e, 0x6d, 0xfc,
	0x22, 0x92, 0xed, 0xab, 0xa5, 0x78, 0xf0, 0xf6, 0xcc, 0x3c, 0xcf, 0xfc, 0x92, 0x99, 0x04, 0x54,
	0x37, 0xa5, 0x56, 0xca, 0x99, 0x60, 0x58, 0x8d, 0x69, 0x48, 0xfc, 0xc2, 0x8f, 0x89, 0xfe, 0x3a,
	0xa2, 0x62, 0x95, 0x7b, 0x96, 0xcf, 0x36, 0xdd, 0x88, 0x45, 0xac, 0x2b, 0x13, 0x5e, 0x1e, 0xca,
	0x4a, 0x16, 0x52, 0xed, 0x27, 0xcd, 0x29, 0x3c, 0x72, 0x48, 0x44, 0x33, 0x41, 0xb8, 0x43, 0xbe,
	0xe4, 0x24, 0x13, 0xf8, 0x39, 0xb4, 0x32, 0xe6, 0xaf, 0x89, 0xe8, 0x07, 0x01, 0x27, 0x59, 0xd6,
	0x41, 0xcf, 0xd0, 0x2b, 0xd5, 0xf9, 0xbb, 0x89, 0x31, 0xd4, 0x12, 0x77, 0x43, 0x3a, 0x77, 0xd2,
	0x94, 0xda, 0x7c, 0x01, 0xad, 0x33, 0x2c, 0x8d, 0x0b, 0xfc, 0x18, 0xea, 0x84, 0x73, 0xc6, 0x0f,
	0x88, 0x7d, 0x61, 0x2e, 0xa1, 0x3e, 0xfe, 0x4a, 0x12, 0x71, 0x62, 0xa0, 0x33, 0x03, 0xbf, 0x81,
	0xa6, 0x1f, 0x71, 0x96, 0xa7, 0x9f, 0x69, 0x12, 0x32, 0x89, 0x6f, 0xda, 0x6d, 0xeb, 0xb4, 0xa0,
	0x35, 0x94, 0xee, 0x24, 0x09, 0x99, 0x03, 0xfe, 0x49, 0x9b, 0x1f, 0x01, 0x24, 0xf4, 0x1f, 0x0f,
	0xfe, 0x6f, 0x76, 0x01, 0x70, 0x76, 0xb0, 0x05, 0xb5, 0x35, 0x4d, 0x02, 0x89, 0x7e, 0x68, 0xeb,
	0x37, 0xc7, 0xad, 0x29, 0x4d, 0x02, 0x47, 0xe6, 0xaa, 0x2d, 0x53, 0x57, 0xac, 0x8e, 0x97, 0xaa,
	0xb4, 0xf9, 0x12, 0x6a, 0x55, 0x02, 0xdf, 0x87, 0x7b, 0xef, 0x17, 0x4b, 0x4d, 0xa9, 0xc4, 0xbb,
	0xc5, 0x48, 0x43, 0xb8, 0x05, 0xea, 0x70, 0x31, 0xff, 0xd0, 0x9f, 0xcc, 0xc7, 0x8e, 0x76, 0x67,
	0x7f, 0x82, 0x27, 0xd3, 0xdc, 0x23, 0x31, 0x11, 0x72, 0xbb, 0x11, 0xcd, 0x52, 0x57, 0xf8, 0x2b,
	0xc2, 0xf1, 0x00, 0x1e, 0x1c, 0x8f, 0x8d, 0x2f, 0x5f, 0xe2, 0xea, 0x73, 0xea, 0x9d, 0x9b, 0x5e,
	0x1a, 0x17, 0xa6, 0x62, 0xcf, 0xa0, 0x3d, 0x3b, 0x9a, 0x92, 0xff, 0xd6, 0x4d, 0x82, 0x98, 0x70,
	0xdc, 0x83, 0xc6, 0x9c, 0x09, 0x1a, 0x16, 0x58, 0xbb, 0x18, 0x97, 0x11, 0xbd, 0x7d, 0xdd, 0x39,
	0xd0, 0x06, 0x4f, 0xb7, 0x3b, 0x03, 0xfd, 0xdc, 0x19, 0xca, 0xb7, 0xd2, 0x40, 0xdb, 0xd2, 0x40,
	0x3f, 0x4a, 0x03, 0xfd, 0x2a, 0x0d, 0xf4, 0xfd, 0xb7, 0xa1, 0x78, 0x0d, 0xf9, 0xc3, 0xf5, 0xfe,
	0x04, 0x00, 0x00, 0xff, 0xff, 0xb0, 0xc0, 0xfa, 0x96, 0xb7, 0x02, 0x00, 0x00,
}
