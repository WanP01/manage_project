package tracing

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"project-user/config"
)

// 固定套路
func JaegerTraceProvider() (*sdktrace.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.AppConf.JaegerC.Endpoints))) //实现otel的接口
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("project-project"), //服务名称
			semconv.DeploymentEnvironmentKey.String("dev"),   //开发环境
		)),
	) // 转成otel 的 provider 类型
	return tp, nil
}
