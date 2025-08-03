package internal_test

import (
	"bufio"
	"manage2/internal"
	"reflect"
	"testing"
)

func TestSliceToExcludesNil(t *testing.T) {
	tests := []struct {
		name   string
		values []any
		want   []any
	}{
		{"empty", []any{}, []any{}},
		{"oneNil", []any{nil}, []any{}},
		{"oneValue", []any{1}, []any{1}},
		{"multiple", []any{1, nil, 2, nil, 3}, []any{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := internal.SliceTo[any](tt.values...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected '%v', received %v", tt.want, got)
			}
		})
	}
}

type hasMessage interface {
	Read(b []byte) (n int, err error)
}

type Messenger struct {
	message string
}

func (t Messenger) Read(b []byte) (n int, err error) {
	copy(b, t.message)
	return len(t.message), nil
}

func TestSliceToAny(t *testing.T) {
	var null hasMessage
	var byIface hasMessage = Messenger{"interface\n"}
	byType := Messenger{"type\n"}

	result := internal.SliceTo[any](1, 2, nil, null, byIface, byType)

	if len(result) != 4 {
		t.Errorf("expected 4 elements, received %d", len(result))
	}
	if result[0] != 1 {
		t.Errorf("expected first element to be 1, received %v", result[0])
	}
	if result[1] != 2 {
		t.Errorf("expected second element to be 2, received %v", result[1])
	}
	if result[2] != byIface {
		t.Errorf("expected third element to be instanceByIface, received %v", result[2])
	}
	if result[3] != byType {
		t.Errorf("expected fourth element to be instanceByType, received %v", result[3])
	}

	checkMessage(t, result[2].(hasMessage), "interface")
	checkMessage(t, result[3].(hasMessage), "type")
}

func checkMessage(t *testing.T, m hasMessage, expected string) {
	t.Helper()
	reader := bufio.NewReader(m)
	msg, isPrefix, err := reader.ReadLine()
	if err != nil {
		t.Errorf("expected no error reading message, received %v", err)
	}
	if isPrefix {
		t.Errorf("expected no prefix, received %v", isPrefix)
	}
	if string(msg) != expected {
		t.Errorf("expected message to be '%s', received %s", expected, string(msg))
	}
}

func TestSliceToMessenger(t *testing.T) {
	var null hasMessage
	var byIface hasMessage = Messenger{"interface\n"}
	byType := Messenger{"type\n"}

	result := internal.SliceTo[Messenger](1, 2, nil, null, byIface, byType)

	if len(result) != 2 {
		t.Errorf("expected 2 elements, received %d", len(result))
	}
	if result[0] != byIface {
		t.Errorf("expected first element to be instanceByIface, received %v", result[0])
	}
	if result[1] != byType {
		t.Errorf("expected second element to be instanceByType, received %v", result[1])
	}

	checkMessage(t, result[0], "interface")
	checkMessage(t, result[1], "type")
}
