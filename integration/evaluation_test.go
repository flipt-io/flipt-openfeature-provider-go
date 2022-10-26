package integration_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/flipt-io/openfeature-provider-go/pkg/provider/flipt"
	"github.com/open-feature/go-sdk/pkg/openfeature"

	"github.com/cucumber/godog"
)

// ctxStorageKey is the key used to pass test data across context.Context
type ctxStorageKey struct{}

var client *openfeature.Client

func aBooleanFlagWithKeyIsEvaluatedWithDefaultValue(
	ctx context.Context, flagKey, defaultValueStr string,
) (context.Context, error) {
	defaultValue, err := strconv.ParseBool(defaultValueStr)
	if err != nil {
		return ctx, errors.New("default value must be of type bool")
	}

	got, err := client.BooleanValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedBooleanValueShouldBe(ctx context.Context, expectedValueStr string) error {
	expectedValue, err := strconv.ParseBool(expectedValueStr)
	if err != nil {
		return errors.New("expected value must be of type bool")
	}

	got, ok := ctx.Value(ctxStorageKey{}).(bool)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got != expectedValue {
		return fmt.Errorf("expected resolved boolean value to be %t, got %t", expectedValue, got)
	}

	return nil
}

func aStringFlagWithKeyIsEvaluatedWithDefaultValue(
	ctx context.Context, flagKey, defaultValue string,
) (context.Context, error) {
	got, err := client.StringValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedStringValueShouldBe(ctx context.Context, expectedValue string) error {
	got, ok := ctx.Value(ctxStorageKey{}).(string)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got != expectedValue {
		return fmt.Errorf("expected resolved string value to be %s, got %s", expectedValue, got)
	}

	return nil
}

func anIntegerFlagWithKeyIsEvaluatedWithDefaultValue(
	ctx context.Context, flagKey string, defaultValue int64,
) (context.Context, error) {
	got, err := client.IntValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedIntegerValueShouldBe(ctx context.Context, expectedValue int64) error {
	got, ok := ctx.Value(ctxStorageKey{}).(int64)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got != expectedValue {
		return fmt.Errorf("expected resolved int value to be %d, got %d", expectedValue, got)
	}

	return nil
}

func aFloatFlagWithKeyIsEvaluatedWithDefaultValue(
	ctx context.Context, flagKey string, defaultValue float64,
) (context.Context, error) {
	got, err := client.FloatValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedFloatValueShouldBe(ctx context.Context, expectedValue float64) error {
	got, ok := ctx.Value(ctxStorageKey{}).(float64)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got != expectedValue {
		return fmt.Errorf("expected resolved int value to be %f, got %f", expectedValue, got)
	}

	return nil
}

func anObjectFlagWithKeyIsEvaluatedWithANullDefaultValue(ctx context.Context, flagKey string) (context.Context, error) {
	got, err := client.ObjectValue(ctx, flagKey, nil, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedObjectValueShouldBeContainFieldsAndWithValuesAndRespectively(
	ctx context.Context, field1, field2, field3, value1, value2 string, value3 int,
) error {
	got, ok := ctx.Value(ctxStorageKey{}).(map[string]interface{})
	if !ok {
		return errors.New("no flag resolution result")
	}

	if err := compareValueToPotentialBool(got[field1], value1); err != nil {
		return fmt.Errorf("field '%s': %w", field1, err)
	}

	if err := compareValueToPotentialBool(got[field2], value2); err != nil {
		return fmt.Errorf("field '%s': %w", field2, err)
	}

	v, ok := got[field3].(float64)
	if !ok {
		return fmt.Errorf("field '%s': expected value to be of type float64, got %T", field3, got[field3])
	}

	if int(v) != value3 {
		return fmt.Errorf(
			"field '%s' expected to contain %d, got %v",
			field3, value3, v,
		)
	}

	return nil
}

func aBooleanFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue(
	ctx context.Context, flagKey string, defaultValueStr string,
) (context.Context, error) {
	defaultValue, err := strconv.ParseBool(defaultValueStr)
	if err != nil {
		return ctx, errors.New("default value must be of type bool")
	}

	got, err := client.BooleanValueDetails(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedBooleanDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe(
	ctx context.Context, valueStr, variant, reason string,
) error {
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return errors.New("value must be of type bool")
	}

	got, ok := ctx.Value(ctxStorageKey{}).(openfeature.BooleanEvaluationDetails)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got.Value != value {
		return fmt.Errorf("expected value to be %t, got %t", value, got.Value)
	}
	// TODO: return variant name (see: https://github.com/flipt-io/flipt/discussions/1075)
	//
	// if got.Variant != variant {
	// 	return fmt.Errorf("expected variant to be %s, got %s", variant, got.Variant)
	// }

	// TODO: We return TARGETING_MATCH here, however flagd returns DEFAULT
	// currently am not sure which one is correct, so need to follow up with the open-feature team
	//
	// if string(got.Reason) != reason {
	// 	return fmt.Errorf("expected reason to be %s, got %s", reason, got.Reason)
	// }

	return nil
}

func aStringFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue(
	ctx context.Context, flagKey, defaultValue string,
) (context.Context, error) {
	got, err := client.StringValueDetails(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedStringDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe(
	ctx context.Context, value, variant, reason string,
) error {
	got, ok := ctx.Value(ctxStorageKey{}).(openfeature.StringEvaluationDetails)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got.Value != value {
		return fmt.Errorf("expected value to be %s, got %s", value, got.Value)
	}
	// TODO: return variant name (see: https://github.com/flipt-io/flipt/discussions/1075)
	//
	// if got.Variant != variant {
	// 	return fmt.Errorf("expected variant to be %s, got %s", variant, got.Variant)
	// }

	// TODO: We return TARGETING_MATCH here, however flagd returns DEFAULT
	// currently am not sure which one is correct, so need to follow up with the open-feature team
	//
	// if string(got.Reason) != reason {
	// 	return fmt.Errorf("expected reason to be %s, got %s", reason, got.Reason)
	// }

	return nil
}

func anIntegerFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue(
	ctx context.Context, flagKey string, defaultValue int64,
) (context.Context, error) {
	got, err := client.IntValueDetails(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedIntegerDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe(
	ctx context.Context, value int64, variant, reason string,
) error {
	got, ok := ctx.Value(ctxStorageKey{}).(openfeature.IntEvaluationDetails)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got.Value != value {
		return fmt.Errorf("expected value to be %d, got %d", value, got.Value)
	}
	// TODO: return variant name (see: https://github.com/flipt-io/flipt/discussions/1075)
	//
	// if got.Variant != variant {
	// 	return fmt.Errorf("expected variant to be %s, got %s", variant, got.Variant)
	// }

	// TODO: We return TARGETING_MATCH here, however flagd returns DEFAULT
	// currently am not sure which one is correct, so need to follow up with the open-feature team
	//
	// if string(got.Reason) != reason {
	// 	return fmt.Errorf("expected reason to be %s, got %s", reason, got.Reason)
	// }

	return nil
}

func aFloatFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue(
	ctx context.Context, flagKey string, defaultValue float64,
) (context.Context, error) {
	got, err := client.FloatValueDetails(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedFloatDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe(
	ctx context.Context, value float64, variant, reason string,
) error {
	got, ok := ctx.Value(ctxStorageKey{}).(openfeature.FloatEvaluationDetails)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got.Value != value {
		return fmt.Errorf("expected value to be %f, got %f", value, got.Value)
	}
	// TODO: return variant name (see: https://github.com/flipt-io/flipt/discussions/1075)
	//
	// if got.Variant != variant {
	// 	return fmt.Errorf("expected variant to be %s, got %s", variant, got.Variant)
	// }

	// TODO: We return TARGETING_MATCH here, however flagd returns DEFAULT
	// currently am not sure which one is correct, so need to follow up with the open-feature team
	//
	// if string(got.Reason) != reason {
	// 	return fmt.Errorf("expected reason to be %s, got %s", reason, got.Reason)
	// }

	return nil
}

func anObjectFlagWithKeyIsEvaluatedWithDetailsAndANullDefaultValue(
	ctx context.Context, flagKey string,
) (context.Context, error) {
	got, err := client.ObjectValueDetails(ctx, flagKey, nil, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}

	return context.WithValue(ctx, ctxStorageKey{}, got), nil
}

func theResolvedObjectDetailsValueShouldBeContainFieldsAndWithValuesAndRespectively(
	ctx context.Context, field1, field2, field3, value1, value2 string, value3 int,
) (context.Context, error) {
	gotResDetail, ok := ctx.Value(ctxStorageKey{}).(openfeature.InterfaceEvaluationDetails)
	if !ok {
		return ctx, errors.New("no flag resolution result")
	}

	got, ok := gotResDetail.Value.(map[string]interface{})
	if !ok {
		return ctx, fmt.Errorf(
			"expected object detail value to be of type map[string]interface{}, got type: %T",
			gotResDetail.Value,
		)
	}

	if err := compareValueToPotentialBool(got[field1], value1); err != nil {
		return ctx, fmt.Errorf("field '%s': %w", field1, err)
	}

	if err := compareValueToPotentialBool(got[field2], value2); err != nil {
		return ctx, fmt.Errorf("field '%s': %w", field2, err)
	}

	v, ok := got[field3].(float64)
	if !ok {
		return ctx, fmt.Errorf("field '%s': expected value to be of type float64, got type: %T", field3, got[field3])
	}

	if int(v) != value3 {
		return ctx, fmt.Errorf(
			"field '%s' expected to contain %d, got %v",
			field3, value3, v)
	}

	return ctx, nil
}

func theVariantShouldBeAndTheReasonShouldBe(ctx context.Context, variant, reason string) error {
	got, ok := ctx.Value(ctxStorageKey{}).(openfeature.InterfaceEvaluationDetails)
	if !ok {
		return errors.New("no flag resolution result")
	}

	if got.Variant != variant {
		return fmt.Errorf("expected variant to be %s, got %s", variant, got.Variant)
	}

	// TODO: We return TARGETING_MATCH here, however flagd returns DEFAULT
	// currently am not sure which one is correct, so need to follow up with the open-feature team
	//
	// if string(got.Reason) != reason {
	// 	return fmt.Errorf("expected reason to be %s, got %s", reason, got.Reason)
	// }

	return nil
}

type contextAwareEvaluationData struct {
	flagKey           string
	defaultValue      string
	evaluationContext openfeature.EvaluationContext
	response          string
}

func contextContainsKeysWithValues(
	ctx context.Context, ctxKey1, ctxKey2, ctxKey3, ctxKey4, ctxValue1, ctxValue2 string, ctxValue3 int64, ctxValue4 string,
) (context.Context, error) {
	evalCtx := openfeature.NewEvaluationContext("test", map[string]interface{}{
		ctxKey1: boolOrString(ctxValue1),
		ctxKey2: boolOrString(ctxValue2),
		ctxKey3: ctxValue3,
		ctxKey4: boolOrString(ctxValue4),
	})

	data := contextAwareEvaluationData{
		evaluationContext: evalCtx,
	}

	return context.WithValue(ctx, ctxStorageKey{}, data), nil
}

func aFlagWithKeyIsEvaluatedWithDefaultValue(
	ctx context.Context, flagKey, defaultValue string,
) (context.Context, error) {
	ctxAwareEvalData, ok := ctx.Value(ctxStorageKey{}).(contextAwareEvaluationData)
	if !ok {
		return ctx, errors.New("no contextAwareEvaluationData found")
	}
	got, err := client.StringValue(ctx, flagKey, defaultValue, ctxAwareEvalData.evaluationContext)
	if err != nil {
		return ctx, fmt.Errorf("openfeature client: %w", err)
	}
	ctxAwareEvalData.flagKey = flagKey
	ctxAwareEvalData.defaultValue = defaultValue
	ctxAwareEvalData.response = got

	return context.WithValue(ctx, ctxStorageKey{}, ctxAwareEvalData), nil
}

func theResolvedStringResponseShouldBe(ctx context.Context, expectedResponse string) (context.Context, error) {
	ctxAwareEvalData, ok := ctx.Value(ctxStorageKey{}).(contextAwareEvaluationData)
	if !ok {
		return ctx, errors.New("no contextAwareEvaluationData found")
	}

	if strings.ToLower(ctxAwareEvalData.response) != strings.ToLower(expectedResponse) {
		return ctx, fmt.Errorf("expected response of '%s', got '%s'", expectedResponse, ctxAwareEvalData.response)
	}

	return ctx, nil
}

func theResolvedFlagValueIsWhenTheContextIsEmpty(ctx context.Context, expectedResponse string) error {
	ctxAwareEvalData, ok := ctx.Value(ctxStorageKey{}).(contextAwareEvaluationData)
	if !ok {
		return errors.New("no contextAwareEvaluationData found")
	}

	got, err := client.StringValue(
		ctx, ctxAwareEvalData.flagKey, ctxAwareEvalData.defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{}))
	if err != nil {
		return fmt.Errorf("openfeature client: %w", err)
	}

	if got != expectedResponse {
		return fmt.Errorf("expected response of '%s', got '%s'", expectedResponse, got)
	}

	return nil
}

type stringFlagNotFoundData struct {
	evalDetails  openfeature.StringEvaluationDetails
	defaultValue string
	err          error
}

func aNonexistentStringFlagWithKeyIsEvaluatedWithDetailsAndADefaultValue(
	ctx context.Context, flagKey, defaultValue string,
) (context.Context, error) {
	got, err := client.StringValueDetails(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{}))

	return context.WithValue(ctx, ctxStorageKey{}, stringFlagNotFoundData{
		evalDetails:  got,
		defaultValue: defaultValue,
		err:          err,
	}), nil
}

func thenTheDefaultStringValueShouldBeReturned(ctx context.Context) (context.Context, error) {
	strNotFoundData, ok := ctx.Value(ctxStorageKey{}).(stringFlagNotFoundData)
	if !ok {
		return ctx, errors.New("no stringFlagNotFoundData found")
	}

	if strNotFoundData.evalDetails.Value != strNotFoundData.defaultValue {
		return ctx, fmt.Errorf(
			"expected default value '%s', got '%s'",
			strNotFoundData.defaultValue, strNotFoundData.evalDetails.Value,
		)
	}

	return ctx, nil
}

func theReasonShouldIndicateAnErrorAndTheErrorCodeShouldIndicateAMissingFlagWith(
	ctx context.Context, errorCode string,
) error {
	strNotFoundData, ok := ctx.Value(ctxStorageKey{}).(stringFlagNotFoundData)
	if !ok {
		return errors.New("no stringFlagNotFoundData found")
	}

	if strNotFoundData.evalDetails.Reason != openfeature.ErrorReason {
		return fmt.Errorf(
			"expected reason '%s', got '%s'",
			openfeature.ErrorReason, strNotFoundData.evalDetails.Reason,
		)
	}

	if string(strNotFoundData.evalDetails.ErrorCode) != errorCode {
		return fmt.Errorf(
			"expected error code '%s', got '%s'",
			errorCode, strNotFoundData.evalDetails.ErrorCode,
		)
	}

	if strNotFoundData.err == nil {
		return errors.New("expected flag evaluation to return an error, got nil")
	}

	return nil
}

type typeErrorData struct {
	evalDetails  openfeature.IntEvaluationDetails
	defaultValue int64
	err          error
}

func aStringFlagWithKeyIsEvaluatedAsAnIntegerWithDetailsAndADefaultValue(
	ctx context.Context, flagKey string, defaultValue int64,
) (context.Context, error) {
	got, err := client.IntValueDetails(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("test", map[string]interface{}{
		"evaluate": true,
	}))

	return context.WithValue(ctx, ctxStorageKey{}, typeErrorData{
		evalDetails:  got,
		defaultValue: defaultValue,
		err:          err,
	}), nil
}

func thenTheDefaultIntegerValueShouldBeReturned(ctx context.Context) (context.Context, error) {
	typeErrData, ok := ctx.Value(ctxStorageKey{}).(typeErrorData)
	if !ok {
		return ctx, errors.New("no typeErrorData found")
	}

	if typeErrData.evalDetails.Value != typeErrData.defaultValue {
		return ctx, fmt.Errorf(
			"expected default value %d, got %d",
			typeErrData.defaultValue, typeErrData.evalDetails.Value,
		)
	}

	return ctx, nil
}

func theReasonShouldIndicateAnErrorAndTheErrorCodeShouldIndicateATypeMismatchWith(
	ctx context.Context, expectedErrorCode string,
) error {
	typeErrData, ok := ctx.Value(ctxStorageKey{}).(typeErrorData)
	if !ok {
		return errors.New("no typeErrorData found")
	}

	if typeErrData.evalDetails.Reason != openfeature.ErrorReason {
		return fmt.Errorf(
			"expected reason '%s', got '%s'",
			openfeature.ErrorReason, typeErrData.evalDetails.Reason,
		)
	}

	if typeErrData.evalDetails.ErrorCode != openfeature.TypeMismatchCode {
		return fmt.Errorf(
			"expected error code '%s', got '%s'",
			openfeature.TypeMismatchCode, typeErrData.evalDetails.ErrorCode,
		)
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a boolean flag with key "([^"]*)" is evaluated with default value "([^"]*)"$`, aBooleanFlagWithKeyIsEvaluatedWithDefaultValue)
	ctx.Step(`^the resolved boolean value should be "([^"]*)"$`, theResolvedBooleanValueShouldBe)

	ctx.Step(`^a string flag with key "([^"]*)" is evaluated with default value "([^"]*)"$`, aStringFlagWithKeyIsEvaluatedWithDefaultValue)
	ctx.Step(`^the resolved string value should be "([^"]*)"$`, theResolvedStringValueShouldBe)

	ctx.Step(`^an integer flag with key "([^"]*)" is evaluated with default value (\d+)$`, anIntegerFlagWithKeyIsEvaluatedWithDefaultValue)
	ctx.Step(`^the resolved integer value should be (\d+)$`, theResolvedIntegerValueShouldBe)

	ctx.Step(`^a float flag with key "([^"]*)" is evaluated with default value (\-*\d+\.\d+)$`, aFloatFlagWithKeyIsEvaluatedWithDefaultValue)
	ctx.Step(`^the resolved float value should be (\-*\d+\.\d+)$`, theResolvedFloatValueShouldBe)

	ctx.Step(`^an object flag with key "([^"]*)" is evaluated with a null default value$`, anObjectFlagWithKeyIsEvaluatedWithANullDefaultValue)
	ctx.Step(`^the resolved object value should be contain fields "([^"]*)", "([^"]*)", and "([^"]*)", with values "([^"]*)", "([^"]*)" and (\d+), respectively$`, theResolvedObjectValueShouldBeContainFieldsAndWithValuesAndRespectively)

	ctx.Step(`^a boolean flag with key "([^"]*)" is evaluated with details and default value "([^"]*)"$`, aBooleanFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue)
	ctx.Step(`^the resolved boolean details value should be "([^"]*)", the variant should be "([^"]*)", and the reason should be "([^"]*)"$`, theResolvedBooleanDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe)

	ctx.Step(`^a string flag with key "([^"]*)" is evaluated with details and default value "([^"]*)"$`, aStringFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue)
	ctx.Step(`^the resolved string details value should be "([^"]*)", the variant should be "([^"]*)", and the reason should be "([^"]*)"$`, theResolvedStringDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe)

	ctx.Step(`^an integer flag with key "([^"]*)" is evaluated with details and default value (\d+)$`, anIntegerFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue)
	ctx.Step(`^the resolved integer details value should be (\d+), the variant should be "([^"]*)", and the reason should be "([^"]*)"$`, theResolvedIntegerDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe)

	ctx.Step(`^a float flag with key "([^"]*)" is evaluated with details and default value (\-*\d+\.\d+)$`, aFloatFlagWithKeyIsEvaluatedWithDetailsAndDefaultValue)
	ctx.Step(`^the resolved float details value should be (\-*\d+\.\d+), the variant should be "([^"]*)", and the reason should be "([^"]*)"$`, theResolvedFloatDetailsValueShouldBeTheVariantShouldBeAndTheReasonShouldBe)

	ctx.Step(`^an object flag with key "([^"]*)" is evaluated with details and a null default value$`, anObjectFlagWithKeyIsEvaluatedWithDetailsAndANullDefaultValue)
	ctx.Step(`^the resolved object details value should be contain fields "([^"]*)", "([^"]*)", and "([^"]*)", with values "([^"]*)", "([^"]*)" and (\d+), respectively$`, theResolvedObjectDetailsValueShouldBeContainFieldsAndWithValuesAndRespectively)
	ctx.Step(`^the variant should be "([^"]*)", and the reason should be "([^"]*)"$`, theVariantShouldBeAndTheReasonShouldBe)

	ctx.Step(`^context contains keys "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" with values "([^"]*)", "([^"]*)", (\d+), "([^"]*)"$`, contextContainsKeysWithValues)
	ctx.Step(`^a flag with key "([^"]*)" is evaluated with default value "([^"]*)"$`, aFlagWithKeyIsEvaluatedWithDefaultValue)
	ctx.Step(`^the resolved string response should be "([^"]*)"$`, theResolvedStringResponseShouldBe)
	ctx.Step(`^the resolved flag value is "([^"]*)" when the context is empty$`, theResolvedFlagValueIsWhenTheContextIsEmpty)

	ctx.Step(`^a non-existent string flag with key "([^"]*)" is evaluated with details and a default value "([^"]*)"$`, aNonexistentStringFlagWithKeyIsEvaluatedWithDetailsAndADefaultValue)
	ctx.Step(`^then the default string value should be returned$`, thenTheDefaultStringValueShouldBeReturned)
	ctx.Step(`^the reason should indicate an error and the error code should indicate a missing flag with "([^"]*)"$`, theReasonShouldIndicateAnErrorAndTheErrorCodeShouldIndicateAMissingFlagWith)

	ctx.Step(`^a string flag with key "([^"]*)" is evaluated as an integer, with details and a default value (\d+)$`, aStringFlagWithKeyIsEvaluatedAsAnIntegerWithDetailsAndADefaultValue)
	ctx.Step(`^then the default integer value should be returned$`, thenTheDefaultIntegerValueShouldBeReturned)
	ctx.Step(`^the reason should indicate an error and the error code should indicate a type mismatch with "([^"]*)"$`, theReasonShouldIndicateAnErrorAndTheErrorCodeShouldIndicateATypeMismatchWith)
}

func TestFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// register the flipt provider before the tests
	fmt.Println("Setting flipt provider...")
	openfeature.SetProvider(flipt.NewProvider())
	fmt.Println("flipt provider configured!")

	client = openfeature.NewClient("integration tests")

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../test-harness/features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func compareValueToPotentialBool(got interface{}, expected string) error {
	expectedBool, err := strconv.ParseBool(expected)
	if err != nil {
		if got.(string) != expected {
			return fmt.Errorf("expected value to be '%s', got '%s'", expected, got.(string))
		}
	} else {
		if got.(bool) != expectedBool {
			return fmt.Errorf("expected value to be %t, got %t", expectedBool, got.(bool))
		}
	}

	return nil
}

func boolOrString(str string) interface{} {
	boolean, err := strconv.ParseBool(str)
	if err != nil {
		return str
	}

	return boolean
}
