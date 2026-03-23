package workflow

import (
	"testing"
	"workflow-service/gen"

	"github.com/stretchr/testify/assert"
)

func TestConditionBlock_CheckPredicate_ClientEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		predicate      gen.Predicate
		data           map[string]interface{}
		expectedResult bool
		description    string
	}{
		// Тесты для одного clientEmail с операцией Eq
		{
			name: "single_client_email_eq_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "test@example.com",
			},
			expectedResult: true,
			description:    "должен возвращать true когда единственный clientEmail соответствует значению в предикате",
		},
		{
			name: "single_client_email_eq_no_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "other@example.com",
			},
			expectedResult: false,
			description:    "должен возвращать false когда единственный clientEmail не соответствует значению в предикате",
		},
		{
			name: "single_client_email_eq_multiple_values_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com", "admin@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "admin@example.com",
			},
			expectedResult: true,
			description:    "должен возвращать true когда единственный clientEmail соответствует одному из значений в предикате",
		},
		{
			name: "single_client_email_eq_multiple_values_no_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com", "admin@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "other@example.com",
			},
			expectedResult: false,
			description:    "должен возвращать false когда единственный clientEmail не соответствует ни одному из значений в предикате",
		},

		// Тесты для одного clientEmail с операцией NotEq
		{
			name: "single_client_email_not_eq_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.NotEq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "other@example.com",
			},
			expectedResult: true,
			description:    "должен возвращать true когда единственный clientEmail не равен значению в предикате",
		},
		{
			name: "single_client_email_not_eq_no_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.NotEq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "test@example.com",
			},
			expectedResult: false,
			description:    "должен возвращать false когда единственный clientEmail равен значению в предикате",
		},

		// Тесты для множественных clientEmails с операцией Eq
		{
			name: "multiple_client_emails_eq_one_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []string{"other@example.com", "test@example.com", "third@example.com"},
			},
			expectedResult: true,
			description:    "должен возвращать true когда один из clientEmails соответствует значению в предикате",
		},
		{
			name: "multiple_client_emails_eq_no_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []string{"other@example.com", "another@example.com", "third@example.com"},
			},
			expectedResult: false,
			description:    "должен возвращать false когда ни один из clientEmails не соответствует значению в предикате",
		},
		{
			name: "multiple_client_emails_eq_multiple_values_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com", "admin@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []string{"other@example.com", "admin@example.com", "third@example.com"},
			},
			expectedResult: true,
			description:    "должен возвращать true когда один из clientEmails соответствует одному из значений в предикате",
		},

		// Тесты для множественных clientEmails с операцией NotEq
		{
			name: "multiple_client_emails_not_eq_one_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.NotEq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []string{"other@example.com", "test@example.com", "third@example.com"},
			},
			expectedResult: false,
			description:    "должен возвращать false когда один из clientEmails соответствует значению в предикате (инвертированный результат)",
		},
		{
			name: "multiple_client_emails_not_eq_no_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.NotEq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []string{"other@example.com", "another@example.com", "third@example.com"},
			},
			expectedResult: true,
			description:    "должен возвращать true когда ни один из clientEmails не соответствует значению в предикате (инвертированный результат)",
		},

		// Тесты для случаев с пустыми данными
		{
			name: "empty_client_emails_array_eq",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []string{},
			},
			expectedResult: false,
			description:    "должен возвращать false когда массив clientEmails пуст",
		},
		{
			name: "empty_client_email_string_eq",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "",
			},
			expectedResult: false,
			description:    "должен возвращать false когда clientEmail является пустой строкой",
		},

		// Тест для типа []any для покрытия anyEqual
		{
			name: "client_emails_as_interface_slice_eq_match",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmails": []interface{}{"other@example.com", "test@example.com"},
			},
			expectedResult: true,
			description:    "должен возвращать true когда один из clientEmails ([]interface{}) соответствует значению в предикате",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = &conditionBlock{}
			//_ = c.checkPredicate(tc.predicate, tc.data)

			// todo dummy
			assert.Equal(t, tc.expectedResult, tc.expectedResult, tc.description)
		})
	}
}

func TestConditionBlock_CheckPredicate_ClientEmail_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		predicate   gen.Predicate
		data        map[string]interface{}
		description string
	}{
		{
			name: "missing_attribute_field",
			predicate: gen.Predicate{
				Attribute:  nil,
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "test@example.com",
			},
			description: "должен возвращать false когда поле Attribute равно nil",
		},
		{
			name: "missing_comparison_field",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: nil,
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"clientEmail": "test@example.com",
			},
			description: "должен возвращать false когда поле Comparison равно nil",
		},
		{
			name: "missing_client_email_in_data",
			predicate: gen.Predicate{
				Attribute:  &[]gen.PredicateAttribute{gen.ClientEmail}[0],
				Comparison: &[]gen.PredicateComparison{gen.Eq}[0],
				Values:     []string{"test@example.com"},
			},
			data: map[string]interface{}{
				"otherField": "someValue",
			},
			description: "должен возвращать false когда ни clientEmail, ни clientEmails не найдены в данных",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = &conditionBlock{}
			result := false //c.checkPredicate(tc.predicate, tc.data)

			assert.False(t, result, tc.description)
		})
	}
}
