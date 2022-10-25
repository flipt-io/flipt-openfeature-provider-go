package flipt_test

import (
	"context"

	"github.com/flipt-io/openfeature-provider-go/pkg/provider/flipt"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

func Example() {
	openfeature.SetProvider(flipt.NewProvider(
		flipt.WithHost("localhost"),
		flipt.WithPort(9000),
		flipt.WithServiceType(flipt.ServiceTypeGRPC),
	))

	client := openfeature.NewClient("my-app")
	value, err := client.BooleanValue(
		context.Background(), "v2_enabled", false, openfeature.EvaluationContext{
			TargetingKey: "tim@apple.com",
			Attributes: map[string]interface{}{
				"favorite_color": "blue",
			},
		},
	)

	if err != nil {
		panic(err)
	}

	if value {
		// do something
	} else {
		// do something else
	}
}
