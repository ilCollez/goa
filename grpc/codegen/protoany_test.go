package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestProtoAnyGeneration(t *testing.T) {
	root := RunGRPCDSL(t, testdata.ProtoAnyDSL)
	services := CreateGRPCServices(root)

	fs := ProtoFiles("", services)
	require.Len(t, fs, 1, "expected one proto file")

	sections := fs[0].SectionTemplates
	require.GreaterOrEqual(t, len(sections), 3, "expected at least 3 sections")

	code := sectionCode(t, sections...)
	t.Logf("Generated proto code:\n%s", code)

	// Verify that google.protobuf.Any is used in the proto file
	assert.Contains(t, code, "google.protobuf.Any", "proto file should contain google.protobuf.Any")
	assert.Contains(t, code, "import \"google/protobuf/any.proto\"", "proto file should import google/protobuf/any.proto")

	// Verify the fields
	assert.Contains(t, code, "google.protobuf.Any data", "payload should have data field with google.protobuf.Any type")
	assert.Contains(t, code, "google.protobuf.Any response", "result should have response field with google.protobuf.Any type")

	// Try to compile with protoc
	fpath := codegen.CreateTempFile(t, code)
	assert.NoError(t, protoc(defaultProtocCmd, fpath, nil), "proto file should compile successfully")
}

func TestProtoAnyTypeRecognition(t *testing.T) {
	root := RunGRPCDSL(t, testdata.ProtoAnyDSL)

	// Verify that ProtoAny types are recognized in the DSL
	svcExpr := root.API.GRPC.Service("ServiceWithProtoAny")
	require.NotNil(t, svcExpr, "service should exist in DSL")

	methodExpr := svcExpr.Endpoint("MethodWithProtoAny")
	require.NotNil(t, methodExpr, "method should exist in DSL")

	// Check payload has ProtoAny field
	payloadObj := expr.AsObject(methodExpr.MethodExpr.Payload.Type)
	require.NotNil(t, payloadObj, "payload should be an object")

	foundProtoAnyInPayload := false
	for _, attr := range *payloadObj {
		if attr.Attribute.Type.Kind() == expr.ProtoAnyKind {
			foundProtoAnyInPayload = true
			t.Logf("Found ProtoAny in payload field: %s", attr.Name)
		}
	}
	assert.True(t, foundProtoAnyInPayload, "payload should contain ProtoAny field")

	// Check result has ProtoAny field
	resultObj := expr.AsObject(methodExpr.MethodExpr.Result.Type)
	require.NotNil(t, resultObj, "result should be an object")

	foundProtoAnyInResult := false
	for _, attr := range *resultObj {
		if attr.Attribute.Type.Kind() == expr.ProtoAnyKind {
			foundProtoAnyInResult = true
			t.Logf("Found ProtoAny in result field: %s", attr.Name)
		}
	}
	assert.True(t, foundProtoAnyInResult, "result should contain ProtoAny field")
}
