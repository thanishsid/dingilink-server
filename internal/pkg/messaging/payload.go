package messaging

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var typeRegistry = make(map[string]func() any)

// RegisterType registers a type with a string key.
func RegisterType[T any]() {
	var instance T

	typeName := reflect.TypeOf(instance).String() // Using reflect.TypeOf

	typeRegistry[typeName] = func() any {
		return new(T)
	}

	fmt.Println("Registered type:", typeName)
}

// PayloadWrapper is a struct that wraps the payload with its type information.
type PayloadWrapper struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SerializePayload serializes a generic payload along with its type information.
func SerializePayload[T any](payload T) ([]byte, error) {
	// Get the type name of the payload (you can customize this further).
	typeName := fmt.Sprintf("%T", payload)

	// Marshal the payload into JSON.
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Wrap the payload with type information.
	wrapper := PayloadWrapper{
		Type:    typeName,
		Payload: payloadBytes,
	}

	// Serialize the wrapper into JSON.
	return json.Marshal(wrapper)
}

// DeserializePayload deserializes a payload based on registered types.
func DeserializePayload(data []byte) (any, error) {
	var wrapper PayloadWrapper

	// First, deserialize the wrapper to get the type information and raw payload.
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	// Look up the type in the registry.
	typeConstructor, ok := typeRegistry[wrapper.Type]
	if !ok {
		return nil, fmt.Errorf("unknown type: %s", wrapper.Type)
	}

	// Create a new instance of the concrete type.
	target := typeConstructor()

	// Deserialize the raw payload into the concrete type.
	if err := json.Unmarshal(wrapper.Payload, target); err != nil {
		return nil, err
	}

	// Dereference the pointer to get the value (using reflection to ensure it works for any type).
	value := reflect.ValueOf(target).Elem().Interface()

	return value, nil
}
