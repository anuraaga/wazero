package main

import "go.opentelemetry.io/collector/pdata/pcommon"

// Interfaces copied from OTel for testing since internal

type TransformContext interface {
	GetItem() interface{}
	GetInstrumentationScope() pcommon.InstrumentationScope
	GetResource() pcommon.Resource
}

type ExprFunc func(ctx TransformContext) interface{}

type Getter interface {
	Get(ctx TransformContext) interface{}
}

type Setter interface {
	Set(ctx TransformContext, val interface{})
}

type GetSetter interface {
	Getter
	Setter
}
