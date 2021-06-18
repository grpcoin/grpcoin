// Copyright 2021 Ahmet Alp Balkan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: grpcoin.proto

package grpcoin

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TradeAction int32

const (
	TradeAction_UNDEFINED TradeAction = 0
	TradeAction_BUY       TradeAction = 1 // Buy a cryptocurrency using cash holdings.
	TradeAction_SELL      TradeAction = 2 // Sell a cryptocurrency for cash holdings.
)

// Enum value maps for TradeAction.
var (
	TradeAction_name = map[int32]string{
		0: "UNDEFINED",
		1: "BUY",
		2: "SELL",
	}
	TradeAction_value = map[string]int32{
		"UNDEFINED": 0,
		"BUY":       1,
		"SELL":      2,
	}
)

func (x TradeAction) Enum() *TradeAction {
	p := new(TradeAction)
	*p = x
	return p
}

func (x TradeAction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TradeAction) Descriptor() protoreflect.EnumDescriptor {
	return file_grpcoin_proto_enumTypes[0].Descriptor()
}

func (TradeAction) Type() protoreflect.EnumType {
	return &file_grpcoin_proto_enumTypes[0]
}

func (x TradeAction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TradeAction.Descriptor instead.
func (TradeAction) EnumDescriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{0}
}

// QuoteTicker represents a quote request for a coin.
type QuoteTicker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ticker string `protobuf:"bytes,1,opt,name=ticker,proto3" json:"ticker,omitempty"` // e.g. BTC
}

func (x *QuoteTicker) Reset() {
	*x = QuoteTicker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QuoteTicker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuoteTicker) ProtoMessage() {}

func (x *QuoteTicker) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuoteTicker.ProtoReflect.Descriptor instead.
func (*QuoteTicker) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{0}
}

func (x *QuoteTicker) GetTicker() string {
	if x != nil {
		return x.Ticker
	}
	return ""
}

// Quote represents a real-time coin price.
type Quote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	T     *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=t,proto3" json:"t,omitempty"`
	Price *Amount                `protobuf:"bytes,20,opt,name=price,proto3" json:"price,omitempty"`
}

func (x *Quote) Reset() {
	*x = Quote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Quote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Quote) ProtoMessage() {}

func (x *Quote) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Quote.ProtoReflect.Descriptor instead.
func (*Quote) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{1}
}

func (x *Quote) GetT() *timestamppb.Timestamp {
	if x != nil {
		return x.T
	}
	return nil
}

func (x *Quote) GetPrice() *Amount {
	if x != nil {
		return x.Price
	}
	return nil
}

// Amount represents a fractional number precisely without
// as opposed to losing precision due to float representation.
type Amount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The integral part of the amount.
	// For example, 3.50 will have `units`=3.
	Units int64 `protobuf:"varint,1,opt,name=units,proto3" json:"units,omitempty"`
	// Number of nano (10^-9) units that represent the fractional amount.
	// For example, 3.5 is represented as `units`=3 and `nanos`=500,000,000.
	//
	// `nanos` must be between -999,999,999 and +999,999,999 inclusive.
	//
	// If `units` is positive, `nanos` must be positive or zero.
	// If `units` is zero, `nanos` can be positive, zero, or negative.
	// If `units` is negative, `nanos` must be negative or zero.
	// For example, -1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	Nanos int32 `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
}

func (x *Amount) Reset() {
	*x = Amount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Amount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Amount) ProtoMessage() {}

func (x *Amount) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Amount.ProtoReflect.Descriptor instead.
func (*Amount) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{2}
}

func (x *Amount) GetUnits() int64 {
	if x != nil {
		return x.Units
	}
	return 0
}

func (x *Amount) GetNanos() int32 {
	if x != nil {
		return x.Nanos
	}
	return 0
}

type TestAuthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TestAuthRequest) Reset() {
	*x = TestAuthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestAuthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestAuthRequest) ProtoMessage() {}

func (x *TestAuthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestAuthRequest.ProtoReflect.Descriptor instead.
func (*TestAuthRequest) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{3}
}

type TestAuthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId string `protobuf:"bytes,10,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *TestAuthResponse) Reset() {
	*x = TestAuthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestAuthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestAuthResponse) ProtoMessage() {}

func (x *TestAuthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestAuthResponse.ProtoReflect.Descriptor instead.
func (*TestAuthResponse) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{4}
}

func (x *TestAuthResponse) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type PortfolioRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PortfolioRequest) Reset() {
	*x = PortfolioRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortfolioRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortfolioRequest) ProtoMessage() {}

func (x *PortfolioRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortfolioRequest.ProtoReflect.Descriptor instead.
func (*PortfolioRequest) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{5}
}

type PortfolioResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// User's cash holdings in USD.
	CashUsd *Amount `protobuf:"bytes,1,opt,name=cash_usd,json=cashUsd,proto3" json:"cash_usd,omitempty"`
	// User's cryptocurrency positions.
	Positions []*PortfolioPosition `protobuf:"bytes,2,rep,name=positions,proto3" json:"positions,omitempty"`
}

func (x *PortfolioResponse) Reset() {
	*x = PortfolioResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortfolioResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortfolioResponse) ProtoMessage() {}

func (x *PortfolioResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortfolioResponse.ProtoReflect.Descriptor instead.
func (*PortfolioResponse) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{6}
}

func (x *PortfolioResponse) GetCashUsd() *Amount {
	if x != nil {
		return x.CashUsd
	}
	return nil
}

func (x *PortfolioResponse) GetPositions() []*PortfolioPosition {
	if x != nil {
		return x.Positions
	}
	return nil
}

type PortfolioPosition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ticker *PortfolioPosition_Ticker `protobuf:"bytes,1,opt,name=ticker,proto3" json:"ticker,omitempty"`
	Amount *Amount                   `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *PortfolioPosition) Reset() {
	*x = PortfolioPosition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortfolioPosition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortfolioPosition) ProtoMessage() {}

func (x *PortfolioPosition) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortfolioPosition.ProtoReflect.Descriptor instead.
func (*PortfolioPosition) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{7}
}

func (x *PortfolioPosition) GetTicker() *PortfolioPosition_Ticker {
	if x != nil {
		return x.Ticker
	}
	return nil
}

func (x *PortfolioPosition) GetAmount() *Amount {
	if x != nil {
		return x.Amount
	}
	return nil
}

type TradeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Action   TradeAction          `protobuf:"varint,1,opt,name=action,proto3,enum=grpcoin.TradeAction" json:"action,omitempty"`
	Ticker   *TradeRequest_Ticker `protobuf:"bytes,2,opt,name=ticker,proto3" json:"ticker,omitempty"`
	Quantity *Amount              `protobuf:"bytes,3,opt,name=quantity,proto3" json:"quantity,omitempty"`
}

func (x *TradeRequest) Reset() {
	*x = TradeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TradeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TradeRequest) ProtoMessage() {}

func (x *TradeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TradeRequest.ProtoReflect.Descriptor instead.
func (*TradeRequest) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{8}
}

func (x *TradeRequest) GetAction() TradeAction {
	if x != nil {
		return x.Action
	}
	return TradeAction_UNDEFINED
}

func (x *TradeRequest) GetTicker() *TradeRequest_Ticker {
	if x != nil {
		return x.Ticker
	}
	return nil
}

func (x *TradeRequest) GetQuantity() *Amount {
	if x != nil {
		return x.Quantity
	}
	return nil
}

type TradeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	T             *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=t,proto3" json:"t,omitempty"`
	Action        TradeAction            `protobuf:"varint,2,opt,name=action,proto3,enum=grpcoin.TradeAction" json:"action,omitempty"`
	Ticker        *TradeResponse_Ticker  `protobuf:"bytes,5,opt,name=ticker,proto3" json:"ticker,omitempty"`
	Quantity      *Amount                `protobuf:"bytes,3,opt,name=quantity,proto3" json:"quantity,omitempty"`
	ExecutedPrice *Amount                `protobuf:"bytes,4,opt,name=executed_price,json=executedPrice,proto3" json:"executed_price,omitempty"`
}

func (x *TradeResponse) Reset() {
	*x = TradeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TradeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TradeResponse) ProtoMessage() {}

func (x *TradeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TradeResponse.ProtoReflect.Descriptor instead.
func (*TradeResponse) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{9}
}

func (x *TradeResponse) GetT() *timestamppb.Timestamp {
	if x != nil {
		return x.T
	}
	return nil
}

func (x *TradeResponse) GetAction() TradeAction {
	if x != nil {
		return x.Action
	}
	return TradeAction_UNDEFINED
}

func (x *TradeResponse) GetTicker() *TradeResponse_Ticker {
	if x != nil {
		return x.Ticker
	}
	return nil
}

func (x *TradeResponse) GetQuantity() *Amount {
	if x != nil {
		return x.Quantity
	}
	return nil
}

func (x *TradeResponse) GetExecutedPrice() *Amount {
	if x != nil {
		return x.ExecutedPrice
	}
	return nil
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[10]
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
	return file_grpcoin_proto_rawDescGZIP(), []int{10}
}

type PortfolioPosition_Ticker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ticker string `protobuf:"bytes,1,opt,name=ticker,proto3" json:"ticker,omitempty"` // e.g. BTC
}

func (x *PortfolioPosition_Ticker) Reset() {
	*x = PortfolioPosition_Ticker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortfolioPosition_Ticker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortfolioPosition_Ticker) ProtoMessage() {}

func (x *PortfolioPosition_Ticker) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortfolioPosition_Ticker.ProtoReflect.Descriptor instead.
func (*PortfolioPosition_Ticker) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{7, 0}
}

func (x *PortfolioPosition_Ticker) GetTicker() string {
	if x != nil {
		return x.Ticker
	}
	return ""
}

type TradeRequest_Ticker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ticker string `protobuf:"bytes,1,opt,name=ticker,proto3" json:"ticker,omitempty"` // e.g. BTC
}

func (x *TradeRequest_Ticker) Reset() {
	*x = TradeRequest_Ticker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TradeRequest_Ticker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TradeRequest_Ticker) ProtoMessage() {}

func (x *TradeRequest_Ticker) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TradeRequest_Ticker.ProtoReflect.Descriptor instead.
func (*TradeRequest_Ticker) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{8, 0}
}

func (x *TradeRequest_Ticker) GetTicker() string {
	if x != nil {
		return x.Ticker
	}
	return ""
}

type TradeResponse_Ticker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Symbol string `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"` // e.g. BTC
}

func (x *TradeResponse_Ticker) Reset() {
	*x = TradeResponse_Ticker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpcoin_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TradeResponse_Ticker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TradeResponse_Ticker) ProtoMessage() {}

func (x *TradeResponse_Ticker) ProtoReflect() protoreflect.Message {
	mi := &file_grpcoin_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TradeResponse_Ticker.ProtoReflect.Descriptor instead.
func (*TradeResponse_Ticker) Descriptor() ([]byte, []int) {
	return file_grpcoin_proto_rawDescGZIP(), []int{9, 0}
}

func (x *TradeResponse_Ticker) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

var File_grpcoin_proto protoreflect.FileDescriptor

var file_grpcoin_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x25, 0x0a, 0x0b, 0x51, 0x75, 0x6f,
	0x74, 0x65, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69, 0x63, 0x6b,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72,
	0x22, 0x58, 0x0a, 0x05, 0x51, 0x75, 0x6f, 0x74, 0x65, 0x12, 0x28, 0x0a, 0x01, 0x74, 0x18, 0x0a,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x01, 0x74, 0x12, 0x25, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x14, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x41, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x22, 0x34, 0x0a, 0x06, 0x41, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x75, 0x6e, 0x69, 0x74, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x05, 0x75, 0x6e, 0x69, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61,
	0x6e, 0x6f, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6e, 0x61, 0x6e, 0x6f, 0x73,
	0x22, 0x11, 0x0a, 0x0f, 0x54, 0x65, 0x73, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x2b, 0x0a, 0x10, 0x54, 0x65, 0x73, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x22, 0x12, 0x0a, 0x10, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x79, 0x0a, 0x11, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69,
	0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x08, 0x63, 0x61, 0x73,
	0x68, 0x5f, 0x75, 0x73, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x07, 0x63, 0x61,
	0x73, 0x68, 0x55, 0x73, 0x64, 0x12, 0x38, 0x0a, 0x09, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f,
	0x69, 0x6e, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f, 0x50, 0x6f, 0x73, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22,
	0x99, 0x01, 0x0a, 0x11, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f, 0x50, 0x6f, 0x73,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x39, 0x0a, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e,
	0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72,
	0x12, 0x27, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e,
	0x74, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x20, 0x0a, 0x06, 0x54, 0x69, 0x63,
	0x6b, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x22, 0xc1, 0x01, 0x0a, 0x0c,
	0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x06,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x72, 0x61, 0x64, 0x65, 0x41, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x06, 0x74, 0x69,
	0x63, 0x6b, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x2e, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72,
	0x12, 0x2b, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x41, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x1a, 0x20, 0x0a,
	0x06, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x22,
	0xa5, 0x02, 0x0a, 0x0d, 0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x28, 0x0a, 0x01, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x01, 0x74, 0x12, 0x2c, 0x0a, 0x06, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x72, 0x61, 0x64, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x35, 0x0a, 0x06, 0x74, 0x69, 0x63,
	0x6b, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72,
	0x12, 0x2b, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x41, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x36, 0x0a,
	0x0e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x64, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e,
	0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x0d, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x64,
	0x50, 0x72, 0x69, 0x63, 0x65, 0x1a, 0x20, 0x0a, 0x06, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x2a, 0x2f, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x64, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x07,
	0x0a, 0x03, 0x42, 0x55, 0x59, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x53, 0x45, 0x4c, 0x4c, 0x10,
	0x02, 0x32, 0x3f, 0x0a, 0x0a, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x31, 0x0a, 0x05, 0x57, 0x61, 0x74, 0x63, 0x68, 0x12, 0x14, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f,
	0x69, 0x6e, 0x2e, 0x51, 0x75, 0x6f, 0x74, 0x65, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x1a, 0x0e,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x51, 0x75, 0x6f, 0x74, 0x65, 0x22, 0x00,
	0x30, 0x01, 0x32, 0x4c, 0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x41, 0x0a,
	0x08, 0x54, 0x65, 0x73, 0x74, 0x41, 0x75, 0x74, 0x68, 0x12, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x65,
	0x73, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x32, 0x8c, 0x01, 0x0a, 0x0a, 0x50, 0x61, 0x70, 0x65, 0x72, 0x54, 0x72, 0x61, 0x64, 0x65, 0x12,
	0x44, 0x0a, 0x09, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f, 0x12, 0x19, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69,
	0x6e, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x66, 0x6f, 0x6c, 0x69, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x38, 0x0a, 0x05, 0x54, 0x72, 0x61, 0x64, 0x65, 0x12, 0x15,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e, 0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0x2e,
	0x54, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42,
	0x17, 0x5a, 0x0b, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x6f, 0x69, 0x6e, 0xaa, 0x02,
	0x07, 0x47, 0x72, 0x70, 0x43, 0x6f, 0x69, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_grpcoin_proto_rawDescOnce sync.Once
	file_grpcoin_proto_rawDescData = file_grpcoin_proto_rawDesc
)

func file_grpcoin_proto_rawDescGZIP() []byte {
	file_grpcoin_proto_rawDescOnce.Do(func() {
		file_grpcoin_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpcoin_proto_rawDescData)
	})
	return file_grpcoin_proto_rawDescData
}

var file_grpcoin_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_grpcoin_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_grpcoin_proto_goTypes = []interface{}{
	(TradeAction)(0),                 // 0: grpcoin.TradeAction
	(*QuoteTicker)(nil),              // 1: grpcoin.QuoteTicker
	(*Quote)(nil),                    // 2: grpcoin.Quote
	(*Amount)(nil),                   // 3: grpcoin.Amount
	(*TestAuthRequest)(nil),          // 4: grpcoin.TestAuthRequest
	(*TestAuthResponse)(nil),         // 5: grpcoin.TestAuthResponse
	(*PortfolioRequest)(nil),         // 6: grpcoin.PortfolioRequest
	(*PortfolioResponse)(nil),        // 7: grpcoin.PortfolioResponse
	(*PortfolioPosition)(nil),        // 8: grpcoin.PortfolioPosition
	(*TradeRequest)(nil),             // 9: grpcoin.TradeRequest
	(*TradeResponse)(nil),            // 10: grpcoin.TradeResponse
	(*Empty)(nil),                    // 11: grpcoin.Empty
	(*PortfolioPosition_Ticker)(nil), // 12: grpcoin.PortfolioPosition.Ticker
	(*TradeRequest_Ticker)(nil),      // 13: grpcoin.TradeRequest.Ticker
	(*TradeResponse_Ticker)(nil),     // 14: grpcoin.TradeResponse.Ticker
	(*timestamppb.Timestamp)(nil),    // 15: google.protobuf.Timestamp
}
var file_grpcoin_proto_depIdxs = []int32{
	15, // 0: grpcoin.Quote.t:type_name -> google.protobuf.Timestamp
	3,  // 1: grpcoin.Quote.price:type_name -> grpcoin.Amount
	3,  // 2: grpcoin.PortfolioResponse.cash_usd:type_name -> grpcoin.Amount
	8,  // 3: grpcoin.PortfolioResponse.positions:type_name -> grpcoin.PortfolioPosition
	12, // 4: grpcoin.PortfolioPosition.ticker:type_name -> grpcoin.PortfolioPosition.Ticker
	3,  // 5: grpcoin.PortfolioPosition.amount:type_name -> grpcoin.Amount
	0,  // 6: grpcoin.TradeRequest.action:type_name -> grpcoin.TradeAction
	13, // 7: grpcoin.TradeRequest.ticker:type_name -> grpcoin.TradeRequest.Ticker
	3,  // 8: grpcoin.TradeRequest.quantity:type_name -> grpcoin.Amount
	15, // 9: grpcoin.TradeResponse.t:type_name -> google.protobuf.Timestamp
	0,  // 10: grpcoin.TradeResponse.action:type_name -> grpcoin.TradeAction
	14, // 11: grpcoin.TradeResponse.ticker:type_name -> grpcoin.TradeResponse.Ticker
	3,  // 12: grpcoin.TradeResponse.quantity:type_name -> grpcoin.Amount
	3,  // 13: grpcoin.TradeResponse.executed_price:type_name -> grpcoin.Amount
	1,  // 14: grpcoin.TickerInfo.Watch:input_type -> grpcoin.QuoteTicker
	4,  // 15: grpcoin.Account.TestAuth:input_type -> grpcoin.TestAuthRequest
	6,  // 16: grpcoin.PaperTrade.Portfolio:input_type -> grpcoin.PortfolioRequest
	9,  // 17: grpcoin.PaperTrade.Trade:input_type -> grpcoin.TradeRequest
	2,  // 18: grpcoin.TickerInfo.Watch:output_type -> grpcoin.Quote
	5,  // 19: grpcoin.Account.TestAuth:output_type -> grpcoin.TestAuthResponse
	7,  // 20: grpcoin.PaperTrade.Portfolio:output_type -> grpcoin.PortfolioResponse
	10, // 21: grpcoin.PaperTrade.Trade:output_type -> grpcoin.TradeResponse
	18, // [18:22] is the sub-list for method output_type
	14, // [14:18] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_grpcoin_proto_init() }
func file_grpcoin_proto_init() {
	if File_grpcoin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpcoin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QuoteTicker); i {
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
		file_grpcoin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Quote); i {
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
		file_grpcoin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Amount); i {
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
		file_grpcoin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestAuthRequest); i {
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
		file_grpcoin_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestAuthResponse); i {
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
		file_grpcoin_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortfolioRequest); i {
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
		file_grpcoin_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortfolioResponse); i {
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
		file_grpcoin_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortfolioPosition); i {
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
		file_grpcoin_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TradeRequest); i {
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
		file_grpcoin_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TradeResponse); i {
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
		file_grpcoin_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
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
		file_grpcoin_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortfolioPosition_Ticker); i {
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
		file_grpcoin_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TradeRequest_Ticker); i {
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
		file_grpcoin_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TradeResponse_Ticker); i {
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
			RawDescriptor: file_grpcoin_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   3,
		},
		GoTypes:           file_grpcoin_proto_goTypes,
		DependencyIndexes: file_grpcoin_proto_depIdxs,
		EnumInfos:         file_grpcoin_proto_enumTypes,
		MessageInfos:      file_grpcoin_proto_msgTypes,
	}.Build()
	File_grpcoin_proto = out.File
	file_grpcoin_proto_rawDesc = nil
	file_grpcoin_proto_goTypes = nil
	file_grpcoin_proto_depIdxs = nil
}
