package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	trace2 "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
	"strconv"
	"strings"
	"time"
)

/*
https://cfl.corp.100.me/pages/viewpage.action?pageId=20354244

TraceID
Format: {service}^^{随机字符}|{timestamp}
Regexp: ^[\w-]+\^\^[0-9a-fA-F]{32}\|\d{13}$
Sample: csoss-service^^98e160c7c2c01afbebbb6fdb741e3f64|1630993898283
service: 请求源头服务名
随机字符: 32位16进制(全部小写)
timestamp: 13位，毫秒级别
生成规则:先从上游获取，如果存在，在通过正则进行格式校验，不符合则丢弃，重新生成

SpanID
Format: {upstream_service}:{x.y^z}｜{位随机数}
Regexp: ^([\w-]+\:)*\d+([\.|\^]\d+)*\|[0-9a-fA-F]{16}$
Sample: csoss-service:1.2^3.4|1234567890abcdef
upstream_service: 上游调用服务，如果没有则为空
随机字符: 16位16进制(全部小写)
.: RPC调用/子Span生成时增加, 例如 1 -> 1.1
^: 异步子Span生成时增加, 例如 1.1 -> 1.1^1
*/
type ddmcIDGenerator struct {
	service string
	l       log.Logger
}

func New(service string) trace2.IDGenerator {
	return &ddmcIDGenerator{service: service}
}

func (d *ddmcIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	return d.newTraceID(), d.newSpanID()
}

func (d *ddmcIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return d.newSpanID()
	}
	span.SpanContext().SpanID()

	sid := string(span.SpanContext().SpanID())
	arr := strings.Split(sid, "|")
	if len(arr) != 2 || arr[0] == "" || arr[1] == "" {
		return d.newSpanID()
	}

	var child = 0
	var async = false
	if sdkSpan, ok := span.(trace2.ReadOnlySpan); ok {
		child = sdkSpan.ChildSpanCount()

		for _, attr := range sdkSpan.Attributes() {
			if attr.Key == "_internal_async" {
				async = true
			}
		}
	}

	if async {
		return trace.SpanID(fmt.Sprintf("%s^%d|%s", arr[0], child+1, arr[1]))
	} else {
		return trace.SpanID(fmt.Sprintf("%s.%d|%s", arr[0], child+1, arr[1]))
	}
}

func (d *ddmcIDGenerator) newTraceID() trace.TraceID {
	uid := strings.ReplaceAll(uuid.NewString(), "-", "")
	tid := d.service + uid + "|" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	return trace.TraceID(tid)
}

func (d *ddmcIDGenerator) newSpanID() trace.SpanID {
	uid := strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	sid := d.service + ":1|" + uid
	return trace.SpanID(sid)
}
