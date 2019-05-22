package grpcsrv

import (
	"context"
	"encoding/json"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryLoggingClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	// if havnt seted the parameter of timeout for an rpc request then setting it automatically
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
	}

	startTime := time.Now()
	log.Printf("[grpc-client] start a grpc request...")
	defer func() {
		in, _ := json.Marshal(req)
		out, _ := json.Marshal(reply)
		inStr, outStr := string(in), string(out)
		endTime := time.Now()
		dur := int64(time.Since(startTime) / time.Millisecond)
		if dur >= 300 {
			//warn
			log.Print("grpc", method, "in", inStr, "out", outStr, "err", err, "startTime", startTime, "endTime", endTime, "dur/ms", dur)
		} else {
			//debug
			log.Print("grpc", method, "in", inStr, "out", outStr, "err", err, "startTime", startTime, "endTime", endTime, "dur/ms", dur)
		}
	}()
	err = invoker(ctx, method, req, reply, cc, opts...)

	log.Printf("[grpc-client] end a grpc request.")
	return err
}

func UnaryLoggingServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	log.Printf("[grpc-server] start responding to an grpc request...")

	resp, err := handler(ctx, req)

	in, _ := json.Marshal(req)
	out, _ := json.Marshal(resp)
	inStr, outStr := string(in), string(out)
	endTime := time.Now()

	dur := int64(time.Since(startTime) / time.Millisecond)
	if dur >= 300 {
		//warn
		log.Print("grpc", info.FullMethod, "in", inStr, "out", outStr, "err", err, "startTime", startTime, "endTime", endTime, "dur/ms", dur)
	} else {
		//debug
		log.Print("grpc", info.FullMethod, "in", inStr, "out", outStr, "err", err, "startTime", startTime, "endTime", endTime, "dur/ms", dur)
	}

	log.Printf("[grpc-server] end response to an grpc request...")

	return resp, err
}

func UnaryRecoveryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			err = status.Errorf(codes.Internal, "the panic err of grpc services: %v", e)
		}
	}()

	return handler(ctx, req)
}
