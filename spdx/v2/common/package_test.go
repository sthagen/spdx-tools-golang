package common

import "testing"

func TestOriginator_UnmarshalJSON(t *testing.T) {
	tt := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid originator",
			data:    []byte("\"Person: John Doe\""),
			wantErr: false,
		},
		{
			name:    "valid originator with no space",
			data:    []byte("\"Person:John Doe\""),
			wantErr: false,
		},
		{
			name:    "valid originator with no space - organization",
			data:    []byte("\"Organization:SPDX\""),
			wantErr: false,
		},
		{
			name:    "valid originator with email",
			data:    []byte("\"Organization: ExampleCodeInspect (contact@example.com)\""),
			wantErr: false,
		},
		{
			name:    "invalid originator with no type",
			data:    []byte("\"John Doe\""),
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var o Originator
			err := o.UnmarshalJSON(tc.data)
			if (err != nil) != tc.wantErr {
				t.Errorf("Originator.UnmarshalJSON() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestOriginator_MarshalJSON(t *testing.T) {
	type mock struct {
		*Originator
	}
	tt := []struct {
		name    string
		data    Originator
		wantErr bool
	}{
		{
			name: "valid originator",
			data: Originator{
				Originator:     "John Doe",
				OriginatorType: "Person",
			},
			wantErr: false,
		},
		{
			name: "originator with no type",
			data: Originator{
				Originator: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "invalid originator with type but no entity",
			data: Originator{
				OriginatorType: "Person",
			},
			wantErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.data.MarshalJSON()
			if (err != nil) != tc.wantErr {
				t.Errorf("Originator.MarshalJSON() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
